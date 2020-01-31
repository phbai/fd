package acdrive

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/phbai/FreeDrive/types"
	"github.com/phbai/FreeDrive/util"
)

type AcDrive struct {
}

func downloadBlock(blocks []types.Block, index int, file *os.File, isOccupied chan bool, wg *sync.WaitGroup, mutex sync.Mutex) error {
	block := blocks[index]
	offset := util.GetOffset(blocks, uint64(index))

	err, response := util.GetResponse(block.Url)

	if err != nil {
		return err
	}

	content := response[62:]

	mutex.Lock()
	defer mutex.Unlock()
	_, err = file.WriteAt(content, int64(offset))

	<-isOccupied
	log.Printf("分块%d/%d下载完毕\n", index+1, len(blocks))
	wg.Done()
	return nil
}

func (ac *AcDrive) Upload(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return errors.New(filename + "文件不存在")
	}

	err, token := GetUpToken()
	if err != nil {
		return err
	}
	log.Println("upToken:", token)

	params := &types.AcfunUploadImageRequest{
		Token: token,
		Id:    "WU_FILE_0",
		Name:  filename,
		Type:  "image/jpg",
		Size:  "182156",
		Key:   "bfs/album/121d736f4b3aa42cb0cc5fd1ce53001a39e5b84w.jpg",
	}
	err = UploadImage(params)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (ac *AcDrive) Download(url string) error {
	mutex := sync.Mutex{}
	err, metadata := util.GetMetadata(url)

	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s", metadata.Filename)
	if stat, err := os.Stat(path); err == nil {
		existFileHash := util.CalculateSha1(metadata.Filename)
		if existFileHash == metadata.Sha1 && stat.Size() == metadata.Size {
			log.Println("文件已存在, 且与服务器端内容一致")
			return nil
		}
	}

	// 创建文件
	f, err := os.Create(path)

	defer f.Close()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	beforeDownload := time.Now()
	const cocurrentNum = 8
	requests := make(chan bool, cocurrentNum)

	for i := 0; i < len(metadata.Blocks); i++ {
		requests <- true
		wg.Add(1)
		go downloadBlock(metadata.Blocks, i, f, requests, &wg, mutex)
	}
	close(requests)
	wg.Wait()

	timeElapsed := float64(time.Since(beforeDownload)) / float64(time.Second)
	speed := float64(metadata.Size) / float64(timeElapsed)
	log.Printf("%s (%s) 下载完毕, 用时%.2f秒, 平均速度%s/s\n", metadata.Filename, util.FormatSize(metadata.Size), timeElapsed, util.FormatSize(int64(speed)))

	newHash := util.CalculateSha1(metadata.Filename)

	if newHash == metadata.Sha1 {
		log.Println("文件校验通过")
	} else {
		log.Println("文件校验未通过")
	}
	return nil
}

func (ac *AcDrive) Login(username string, password string) error {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("key", "")
	data.Set("captcha", "")

	response, err := http.PostForm("https://id.app.acfun.cn/rest/web/login/signin", data)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	cookies := response.Cookies()

	cookie := types.AcfunLoginCookie{
		AcPasstoken: cookies[0].Value,
		AuthKey:     cookies[1].Value,
		AcUsername:  cookies[2].Value,
		AcPostHint:  cookies[3].Value,
		AcUserImg:   cookies[4].Value,
	}

	res, err := json.MarshalIndent(cookie, "", "  ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile("cookies.json", res, 0644)
	if err != nil {
		return err
	}

	log.Println("登录成功")
	return nil
}

func (ac *AcDrive) Info(url string) error {
	err, metadata := util.GetMetadata(url)

	if err != nil {
		return err
	}

	fmt.Println("文件名:", metadata.Filename)
	fmt.Println("大小:", util.FormatSize(metadata.Size))
	fmt.Println("SHA-1:", metadata.Sha1)
	fmt.Println("上传时间:", util.FormatTime(metadata.Time))
	fmt.Println("分块数:", len(metadata.Blocks))

	for index, block := range metadata.Blocks {
		fmt.Printf("分块%d (%s) URL: %s\n", index+1, util.FormatSize(block.Size), block.Url)
	}
	return nil
}

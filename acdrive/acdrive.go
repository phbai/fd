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

	"github.com/phbai/fd/types"
	"github.com/phbai/fd/util"
)

type AcDrive struct {
}

func (ac *AcDrive) Upload(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return errors.New(filename + "文件不存在")
	}

	err, token := GetUpToken()
	if err != nil {
		return err
	}

	_, fileSize := GetFileSize(filename)
	blocks, err := ReadChunks(filename)

	blockMetadatas := make([]types.Block, len(blocks))

	if err != nil {
		return err
	}

	const cocurrentNum = 8
	requests := make(chan bool, cocurrentNum)
	var wg sync.WaitGroup

	progressBar := util.NewProgressBar(int(fileSize), filename)

	for index, block := range blocks {
		requests <- true
		wg.Add(1)

		go func(index int, block []byte, wg *sync.WaitGroup, isOccupied chan bool) {
			defer wg.Done()
			fullBlock := append(BlockHeader(block), block...)
			blockSha1 := util.CalculateBlockSha1(block)
			fullBlockSha1 := util.CalculateBlockSha1(fullBlock)

			params := &types.AcfunUploadImageRequest{
				Token: token,
				Id:    "WU_FILE_0",
				Name:  fmt.Sprintf("%s.bmp", fullBlockSha1),
				Type:  "image/bmp",
				Size:  fmt.Sprintf("%d", len(fullBlock)),
				Key:   fmt.Sprintf("bfs/album/%s.bmp", fullBlockSha1),
			}
			err, url := UploadBlock(params, fullBlock)
			if err != nil {
				// return err
				fmt.Printf("上传出错了%s", err)
			}

			blockMetadata := types.Block{
				Size: int64(len(block)),
				Url:  url,
				Sha1: blockSha1,
			}

			blockMetadatas[index] = blockMetadata
			progressBar.AddCompletedSize(len(block))
			<-isOccupied
		}(index, block, &wg, requests)
	}

	close(requests)
	wg.Wait()

	now := time.Now()

	metadata := types.Metadata{
		Time:     uint64(now.Unix()),
		Filename: filename,
		Size:     fileSize,
		Sha1:     util.CalculateFileSha1(filename),
		Blocks:   blockMetadatas,
	}

	metadataBytes, _ := json.Marshal(metadata)

	fullMetadataBlock := append(BlockHeader(metadataBytes), metadataBytes...)
	fullMetadataBlockSha1 := util.CalculateBlockSha1(fullMetadataBlock)

	params := &types.AcfunUploadImageRequest{
		Token: token,
		Id:    "WU_FILE_0",
		Name:  fmt.Sprintf("%s.bmp", fullMetadataBlockSha1),
		Type:  "image/bmp",
		Size:  fmt.Sprintf("%d", len(fullMetadataBlock)),
		Key:   fmt.Sprintf("bfs/album/%s.bmp", fullMetadataBlockSha1),
	}

	err, url := UploadBlock(params, fullMetadataBlock)
	if err != nil {
		return errors.New(fmt.Sprintf("元数据上传失败: %s", err))
	}

	fmt.Printf("上传成功，链接👉 %s\n", FormatUrl(url))
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
		existFileHash := util.CalculateFileSha1(metadata.Filename)
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

	const cocurrentNum = 8
	requests := make(chan bool, cocurrentNum)

	progressBar := util.NewProgressBar(int(metadata.Size), metadata.Filename)

	for i := 0; i < len(metadata.Blocks); i++ {
		requests <- true
		wg.Add(1)
		go DownloadBlock(metadata.Blocks, i, f, requests, &wg, mutex, progressBar)
	}
	close(requests)
	wg.Wait()

	newHash := util.CalculateFileSha1(metadata.Filename)

	if newHash == metadata.Sha1 {
		log.Println("文件校验通过")
	} else {
		log.Println("文件校验未通过,请重新下载")
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

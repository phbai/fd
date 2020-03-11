package ali

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/phbai/fd/types"
	"github.com/phbai/fd/util"
)

var client *http.Client

const CHUNK_SIZE = 4 * 1024 * 1024

func init() {
	client = &http.Client{}
}

func BlockHeader(block []byte) []byte {
	bmpHeader := make([]byte, 62)
	bmpHeader = []byte{
		0x42, 0x4D, // BM,
		0x00, 0x00, 0x00, 0x00, // 小端序: 14 + 40 + 8 + len(data)
		0x00, 0x00,
		0x00, 0x00,
		0x3e, 0x00, 0x00, 0x00,
		0x28, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // 小端序: len(data)
		0x01, 0x00, 0x00, 0x00,
		0x01, 0x00,
		0x01, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // 小端序: math.ceil(len(data) / 8))
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00,
	}

	binary.LittleEndian.PutUint32(bmpHeader[2:6], uint32(14+40+8+len(block)))
	binary.LittleEndian.PutUint32(bmpHeader[18:22], uint32(len(block)))
	binary.LittleEndian.PutUint32(bmpHeader[34:38], uint32(math.Ceil(float64(len(block))/8.0)))
	return bmpHeader
}

func ReadChunks(filename string) ([][]byte, error) {
	file, err := os.Open(filename)

	defer file.Close()

	if err != nil {
		return nil, err
	}

	fileStatus, err := file.Stat()

	if err != nil {
		return nil, err
	}

	fileSize := fileStatus.Size()

	chunksNum := int(math.Ceil(float64(fileSize) / float64(4*1024*1024)))

	res := make([][]byte, chunksNum)

	for j := 0; j < chunksNum; j++ {
		if j+1 >= chunksNum {
			res[j] = make([]byte, int(fileSize)-j*CHUNK_SIZE)
		} else {
			res[j] = make([]byte, CHUNK_SIZE)
		}
	}

	for i := 0; i < chunksNum; i++ {
		offset := i * CHUNK_SIZE
		_, err := file.ReadAt(res[i], int64(offset))

		if err == io.EOF {
			fmt.Println("文件读取完毕")
		}
		if err != nil && err != io.EOF {
			return nil, errors.New(fmt.Sprintf("第%d片读取失败\n", i+1))
		}
	}

	return res, nil
}

func UploadBlock(params *types.AliUploadImageRequest, block []byte) (error, string) {
	url := "https://kfupload.alibaba.com/mupload"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	path := params.Name
	part1, errFile1 := writer.CreateFormFile("file", filepath.Base(path))
	_, errFile1 = io.Copy(part1, bytes.NewReader(block))

	if errFile1 != nil {
		return errFile1, ""
	}

	_ = writer.WriteField("name", params.Name)
	_ = writer.WriteField("scene", "aeMessageCenterV2ImageRule")

	err := writer.Close()
	if err != nil {
		return err, ""
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		return err, ""
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "iAliexpress/6.22.1 (iPhone; iOS 12.1.2; Scale/2.00)")
	req.Header.Add("Accept", "application/json")
	// req.Header.Add("Accept-Encoding", "gzip,deflate,sdch")
	req.Header.Add("Connection", "close")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := client.Do(req)

	if err != nil {
		return err, ""
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))

	var result types.AliUploadImageResponse

	err = json.Unmarshal(body, &result)

	if err != nil {
		return err, ""
	}

	if result.Url == "" {
		return errors.New("上传失败"), ""
	}
	return nil, result.Url
}

func GetFileSize(filename string) (error, int64) {
	file, err := os.Open(filename)

	defer file.Close()

	if err != nil {
		return err, 0
	}

	fileStatus, err := file.Stat()

	if err != nil {
		return err, 0
	}

	fileSize := fileStatus.Size()
	return nil, fileSize
}

func FormatUrl(url string) string {
	re := regexp.MustCompile(`[a-zA-Z0-9]{32,}`)
	return fmt.Sprintf("fd03://%s", re.FindString(url))
}

func DownloadBlock(blocks []types.Block, index int, file *os.File, isOccupied chan bool, wg *sync.WaitGroup, mutex sync.Mutex, bar *util.ProgressBar) error {
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
	wg.Done()

	bar.AddCompletedSize(int(block.Size))
	return nil
}

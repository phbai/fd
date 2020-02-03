package acdrive

import (
	"bytes"
	"encoding/base64"
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
	"strings"

	"github.com/phbai/FreeDrive/types"
)

var client *http.Client

const CHUNK_SIZE = 4 * 1024 * 1024

func init() {
	client = &http.Client{}
}

func AddCookie(req *http.Request) error {
	var cookie types.AcfunLoginCookie

	content, err := ioutil.ReadFile("cookies.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, &cookie)

	if err != nil {
		return err
	}

	req.AddCookie(&http.Cookie{Name: "acPasstoken", Value: cookie.AcPasstoken})
	req.AddCookie(&http.Cookie{Name: "auth_key", Value: cookie.AuthKey})
	req.AddCookie(&http.Cookie{Name: "ac_username", Value: cookie.AcUsername})
	req.AddCookie(&http.Cookie{Name: "acPostHint", Value: cookie.AcPostHint})
	req.AddCookie(&http.Cookie{Name: "ac_userimg", Value: cookie.AcUserImg})

	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36")

	return nil
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

func UploadBlock(params *types.AcfunUploadImageRequest, block []byte) (error, string) {
	url := "https://upload.qiniup.com"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	path := params.Name
	part1, errFile1 := writer.CreateFormFile("file", filepath.Base(path))
	_, errFile1 = io.Copy(part1, bytes.NewReader(block))

	if errFile1 != nil {
		return errFile1, ""
	}
	_ = writer.WriteField("token", params.Token)
	_ = writer.WriteField("id", params.Id)
	_ = writer.WriteField("name", params.Name)
	_ = writer.WriteField("type", params.Type)
	_ = writer.WriteField("size", params.Size)
	_ = writer.WriteField("key", params.Key)
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
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := client.Do(req)

	if err != nil {
		return err, ""
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var result types.AcfunUploadImageResponse

	err = json.Unmarshal(body, &result)

	if err != nil {
		return err, ""
	}
	return nil, fmt.Sprintf("https://imgs.aixifan.com/%s", result.Key)
}

func GetUpToken() (error, string) {
	req, err := http.NewRequest("GET", "https://www.acfun.cn/v2/user/content/upToken", nil)

	AddCookie(req)

	if err != nil {
		return err, ""
	}

	req.Header.Add("devicetype", "7")
	// req.Header.Add("Referer", "https://t.bilibili.com/")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return err, ""
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err, ""
	}

	var tokenObject types.AcfunGetToken
	err = json.Unmarshal(body, &tokenObject)

	if err != nil {
		return err, ""
	}

	token, err := base64.URLEncoding.DecodeString(tokenObject.Vdata.Uptoken)

	if err != nil {
		return err, ""
	}

	upToken := strings.Replace(string(token), "null:", "", -1)
	return nil, upToken
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
	re := regexp.MustCompile(`[a-fA-F0-9]{40}`)
	return fmt.Sprintf("adrive://%s", re.FindString(url))
}

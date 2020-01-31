package acdrive

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phbai/FreeDrive/types"
)

var client *http.Client

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

func UploadImage(params *types.AcfunUploadImageRequest) (err error) {
	url := "https://upload.qiniup.com"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	path := params.Name
	file, errFile1 := os.Open(path)
	defer file.Close()
	part1, errFile1 := writer.CreateFormFile("file", filepath.Base(path))
	_, errFile1 = io.Copy(part1, file)

	if errFile1 != nil {
		return errFile1
	}
	_ = writer.WriteField("token", params.Token)
	_ = writer.WriteField("id", params.Id)
	_ = writer.WriteField("name", params.Name)
	_ = writer.WriteField("type", params.Type)
	_ = writer.WriteField("size", params.Size)
	_ = writer.WriteField("key", params.Key)
	err = writer.Close()
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))
	return nil
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

	fmt.Println(string(body))
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

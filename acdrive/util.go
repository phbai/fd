package acdrive

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/phbai/FreeDrive/types"
)

func GetUpToken() (error, string) {
	content, err := ioutil.ReadFile("cookies.json")
	if err != nil {
		return err, ""
	}

	var cookie types.AcfunLoginCookie

	err = json.Unmarshal(content, &cookie)

	if err != nil {
		return err, ""
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://www.acfun.cn/v2/user/content/upToken", nil)

	req.AddCookie(&http.Cookie{ Name: "acPasstoken", Value: cookie.AcPasstoken })
	req.AddCookie(&http.Cookie{ Name: "auth_key", Value: cookie.AuthKey })
	req.AddCookie(&http.Cookie{ Name: "ac_username", Value: cookie.AcUsername })
	req.AddCookie(&http.Cookie{ Name: "acPostHint", Value: cookie.AcPostHint })
	req.AddCookie(&http.Cookie{ Name: "ac_userimg", Value: cookie.AcUserImg })

	if err != nil {
		return err, ""
	}

	req.Header.Add("devicetype", "7")
	// req.Header.Add("Referer", "https://t.bilibili.com/")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36")

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
		return err, "";
	}

	upToken := strings.Replace(string(token), "null:", "", -1)
	return nil, upToken
}
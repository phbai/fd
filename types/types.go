package types

type Block struct {
	Size int64  `json:"size"`
	Url  string `json:"url"`
	Sha1 string `json:"sha1"`
}

type Metadata struct {
	Time     uint64  `json:"time"`
	Filename string  `json:"filename"`
	Size     int64   `json:"size"`
	Sha1     string  `json:"sha1"`
	Blocks   []Block `json:"block"`
}

type AcfunLoginCookie struct {
	AcPasstoken string `json:"acPasstoken"`
	AuthKey     string `json:"auth_key"`
	AcUsername  string `json:"ac_username"`
	AcPostHint  string `json:"acPostHint"`
	AcUserImg   string `json:"ac_userimg"`
}

type AcfunGetToken struct {
	Errorid   int               `json:"errorid"`
	Requestid string            `json:"requestid"`
	Errordesc string            `json:"errordesc,omitempty"`
	Vdata     AcfunGetTokenData `json:"vdata"`
}

type AcfunGetTokenData struct {
	Uptoken string `json:"uptoken"`
	Url     string `json:"url"`
}

type AcfunUploadImageResponse struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

type AcfunUploadImageRequest struct {
	Token string
	Id    string
	Name  string
	Type  string
	Size  string
	Key   string
}

type BaijiahaoUploadImageRequest struct {
	Name string
}

type BaijiahaoUploadImageResponse struct {
	ErrMsg string `json:"errmsg"`
	Ret    BaijiahaoUploadImageResponseReturn
}

type BaijiahaoUploadImageResponseReturn struct {
	OrgUrl string `json:"org_url"`
	Mime   string `json:"mime"`
	Name   string `json:"name"`
	Size   int    `json:"size"`
	Type   string `json:"type"`
}

type AliUploadImageRequest struct {
	Name  string
	Scene string
}

type AliUploadImageResponse struct {
	FsUrl  string `json:"fs_url"`
	Code   string `json:"code"`
	Size   string `json:"size"`
	Width  string `json:"width"`
	Url    string `json:"url"`
	Hash   string `json:"hash"`
	Height string `json:"height"`
}

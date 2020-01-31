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

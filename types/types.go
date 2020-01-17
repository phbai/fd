package types

type Block struct {
	Size int64 `json:"size"`
	Url string `json:"url"`
	Sha1 string `json:"sha1"`
}

type Metadata struct {
	Time uint64 `json:"time"`
	Filename string `json:"filename"`
	Size int64 `json:"size"`
	Sha1 string `json:"sha1"`
	Blocks []Block `json:"block"`
}
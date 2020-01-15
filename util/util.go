package util

import (
	"time"
	"fmt"
	"regexp"
	"errors"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/phbai/FreeDrive/types"
)

func FormatTime(timestamp uint64) string {
	tm := time.Unix(int64(timestamp), 0)
	cstZone := time.FixedZone("GMT", 8*3600)       // 东八
	return tm.In(cstZone).Format("2006-01-02 15:04:05")
}

func FormatSize(filesize uint64) string {
	var unit = [7]string{"B", "KB", "MB", "GB", "TB", "PB", "ZB"}
	var size = float64(filesize);
	count := 0;
	
	for size >= 1024 {
	  size = size / 1024
	  count++
	}
	return fmt.Sprintf("%.2f%s", size, unit[count])
}

func GetResponse(sha1 string) (error, []byte) {
	url := fmt.Sprintf("https://imgs.aixifan.com/bfs/album/%s.bmp", sha1)

	resp, err := http.Get(url)
	if err != nil {
		return err, []byte{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return nil, body
}

func GetMetadata(url string) (error, types.Metadata) {
	re := regexp.MustCompile(`[a-fA-F0-9]{40}`)
	res := re.FindAllString(url, -1)
	if len(res) == 0 {
		return errors.New("输入的地址不合法"), types.Metadata{}
	}

	err, fullMetadata := GetResponse(res[0])
	if err != nil {
		return err, types.Metadata{};
	}
	metadataContent := fullMetadata[62:]

	var metadata types.Metadata
	err = json.Unmarshal(metadataContent, &metadata)
	if err != nil {
		return err, types.Metadata{};
	}
	return nil, metadata;
}

func GetOffset(blocks []types.Block, index uint64) uint64 {
	// offset := 0
	var i, offset uint64 = 0, 0
	i, offset = 0, 0
	for ; i < index; i++ {
		offset += blocks[i].Size
	}
	return offset
}
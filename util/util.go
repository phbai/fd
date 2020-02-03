package util

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/phbai/FreeDrive/types"
)

func FormatTime(timestamp uint64) string {
	tm := time.Unix(int64(timestamp), 0)
	cstZone := time.FixedZone("GMT", 8*3600) // 东八
	return tm.In(cstZone).Format("2006-01-02 15:04:05")
}

func FormatSize(filesize int64) string {
	var unit = [7]string{"B", "KB", "MB", "GB", "TB", "PB", "ZB"}
	var size = float64(filesize)
	count := 0

	for size >= 1024 {
		size = size / 1024
		count++
	}
	return fmt.Sprintf("%.2f%s", size, unit[count])
}

func GetResponse(url string) (error, []byte) {
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

	requestUrl := fmt.Sprintf("https://imgs.aixifan.com/bfs/album/%s.bmp", res[0])
	err, fullMetadata := GetResponse(requestUrl)
	if err != nil {
		return err, types.Metadata{}
	}
	metadataContent := fullMetadata[62:]

	var metadata types.Metadata
	err = json.Unmarshal(metadataContent, &metadata)
	if err != nil {
		return err, types.Metadata{}
	}
	return nil, metadata
}

func GetOffset(blocks []types.Block, index uint64) int64 {
	var i, offset int64 = 0, 0
	for ; i < int64(index); i++ {
		offset += blocks[i].Size
	}
	return offset
}

func CalculateFileSha1(filename string) string {
	f, err := os.Open(filename)
	defer f.Close()

	if err != nil {
		log.Fatal(err)
	}

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func CalculateBlockSha1(block []byte) string {
	h := sha1.New()
	if _, err := io.Copy(h, bytes.NewReader(block)); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

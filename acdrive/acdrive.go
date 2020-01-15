package acdrive

import (
	"fmt"
	"os"
	"github.com/phbai/FreeDrive/util"
)

type AcDrive struct {
}


func downloadBlock(url string, file *os.File, offset uint64, isOccupied chan bool) error {
	isOccupied <- true
	err, response := util.GetResponse(url)

	if err != nil {
		return err;
	}

	content := response[62:]

	n2, err := file.WriteAt(content, int64(offset))
	fmt.Println("写入了%d个字节", n2)
	isFinished <- true
	return nil
}

func (ac *AcDrive) Upload(filename string) {
	fmt.Println("Upload");
}

func (ac *AcDrive) Download(url string) error {
	err, metadata := util.GetMetadata(url)
	
	if err != nil {
		return err
	}

	path := fmt.Sprintf("./%s", metadata.Filename);
	if _, err := os.Stat(path); os.IsExist(err) {
		fmt.Println("当前目录下存在该文件，直接跳过");
		return nil;
	}

	// 创建文件
	f, err := os.Create(path)

	defer f.Close()
	if err != nil {
		return err;
	}
	
	const cocurrentNum = 8;
	requests := make(chan bool, cocurrentNum)

	for i := 0; i < cocurrentNum; i++ {
		requests <- true
	}

	for j := 0; j < len(metadata.Block); j++ {
		go downloadBlock(metadata.Block[j].Url, f, util.GetOffset(metadata.Block, uint64(j)), <- requests)
	}
	return nil
}


func (ac *AcDrive) Login(username string, password string) {
	fmt.Println("Login");
}

func (ac *AcDrive) Info(url string) error {
	err, metadata := util.GetMetadata(url)

	if err != nil {
		return err
	}

	fmt.Println("文件名:", metadata.Filename);
	fmt.Println("大小:", util.FormatSize(metadata.Size));
	fmt.Println("SHA-1:", metadata.Sha1);
	fmt.Println("上传时间:", util.FormatTime(metadata.Time));
	fmt.Println("分块数:", len(metadata.Block));

	for index, block := range metadata.Block {
		fmt.Printf("分块%d (%s) URL: %s\n", index + 1, util.FormatSize(block.Size), block.Url);
	}
	return nil
}

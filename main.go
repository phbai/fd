package main

import (
	"github.com/phbai/fd/baijiahao"
)

func main() {
	baijiahao := &baijiahao.Baijiahao{}
	fd := FreeDrive{baijiahao}
	fd.Run()
}

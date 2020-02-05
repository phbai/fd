package main

import (
	"github.com/phbai/FreeDrive/baijiahao"
)

func main() {
	baijiahao := &baijiahao.Baijiahao{}
	fd := FreeDrive{baijiahao}
	fd.Run()
}

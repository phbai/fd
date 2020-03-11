package main

import (
	"github.com/phbai/fd/ali"
)

func main() {
	ali := &ali.Ali{}
	fd := FreeDrive{ali}
	fd.Run()
}

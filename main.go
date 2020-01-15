package main

import (
	"github.com/phbai/FreeDrive/acdrive"
)

func main() {
	acDrive := &acdrive.AcDrive{};
	fd := FreeDrive{acDrive};
	fd.Run()
}

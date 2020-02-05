package util

import (
	"github.com/gosuri/uiprogress"
)

type ProgressBar struct {
	Size     int
	Filename string
	Bar      *uiprogress.Bar
}

func NewProgressBar(size int, filename string) *ProgressBar {
	uiprogress.Start()
	bar := uiprogress.AddBar(size)
	bar.AppendCompleted()

	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return filename
	})

	return &ProgressBar{
		Size:     size,
		Filename: filename,
		Bar:      bar,
	}
}

func (pb *ProgressBar) AddCompletedSize(completed int) {
	current := pb.Bar.Current()
	pb.Bar.Set(current + completed)
}

package frame

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/frrad/scenesearch/lib/util"
)

type Video struct {
	Filename string
}

func (v *Video) ExtractFrame(offset time.Duration) (io.ReadCloser, error) {
	f, err := ioutil.TempFile(".", "frame*.jpg")
	if err != nil {
		return nil, err
	}

	outName := f.Name()
	log.Print("output will go in ", outName)

	// https://stackoverflow.com/a/27573049/858795

	durationStr := formatDuration(offset)
	log.Println("extracting at offset", durationStr)

	args := []string{
		"-y", // overwrite file
		"-ss", durationStr,
		"-i", v.Filename,
		"-vframes", "1",
		"-q:v ", "2",
		outName,
	}

	util.ExecDebug("ffmpeg", args...)

	return f, nil
}

func formatDuration(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour

	m := d / time.Minute
	d -= m * time.Minute

	s := d / time.Second
	d -= s * time.Second

	ms := d / time.Millisecond

	return fmt.Sprintf("%d:%02d:%02d.%03d", h, m, s, ms)
}

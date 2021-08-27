package frame

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/frrad/scenesearch/lib/util"
)

func (v *Video) frameDoneFileName(offset time.Duration) string {
	return fmt.Sprintf("./%s/%s-%d.jpeg", cacheName, v.Filename, offset)
}

func (v *Video) Frame(offset time.Duration) (string, error) {
	cachedName := v.frameDoneFileName(offset)

	_, err := os.Stat(cachedName)
	if err == nil {
		return cachedName, nil
	}

	err = v.extractFrame(offset, cachedName)
	if err != nil {
		return "", err
	}

	return cachedName, nil
}

// extractFrame extracts a frame from the video by calling ffmpeg
func (v *Video) extractFrame(offset time.Duration, outName string) error {
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

	_, err := util.ExecDebug("ffmpeg", args...)
	return err
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

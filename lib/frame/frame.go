package frame

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/frrad/scenesearch/lib/util"
)

func (v *Video) frameDoneFileName(offset time.Duration) string {
	return fmt.Sprintf("./%s/%s-%d.jpeg", cacheName, v.Filename, offset)
}

func (v *Video) cachedFrame(offset time.Duration) (io.ReadCloser, error) {
	fn := v.frameDoneFileName(offset)

	_, err := os.Stat(fn)
	if err != nil {
		return nil, err
	}

	return os.Open(fn)
}

func (v *Video) Frame(offset time.Duration) (io.ReadCloser, error) {
	ans, err := v.cachedFrame(offset)
	if err == nil {
		return ans, nil
	}

	f, err := v.extractFrame(offset)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	f.Close()

	err = ioutil.WriteFile(v.frameDoneFileName(offset), b, 0755)
	if err != nil {
		return nil, err
	}

	ans, err = v.cachedFrame(offset)
	if err != nil {
		return ans, err
	}

	return ans, nil
}

// extractFrame extracts a frame from the video by calling ffmpeg
func (v *Video) extractFrame(offset time.Duration) (io.ReadCloser, error) {
	f, err := ioutil.TempFile("", "frame*.jpg")
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

	_, err = util.ExecDebug("ffmpeg", args...)
	if err != nil {
		return nil, err
	}

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

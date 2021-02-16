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

func (v *Video) splitDoneFileName(start, end time.Duration) string {
	return fmt.Sprintf("./splitcache/%s-%d-%d.mp4", v.Filename, start, end)
}

func (v *Video) cachedSplit(start, end time.Duration) (io.ReadCloser, error) {
	fn := v.splitDoneFileName(start, end)

	_, err := os.Stat(fn)
	if err != nil {
		return nil, err
	}

	return os.Open(fn)
}

func (v *Video) Split(start, end time.Duration) (io.ReadCloser, error) {
	ans, err := v.cachedSplit(start, end)
	if err == nil {
		return ans, nil
	}

	f, err := v.extractSplit(start, end)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	f.Close()

	err = ioutil.WriteFile(v.splitDoneFileName(start, end), b, 0755)
	if err != nil {
		return nil, err
	}

	ans, err = v.cachedSplit(start, end)
	if err != nil {
		return ans, err
	}

	return ans, nil
}

func (v *Video) extractSplit(startOffset, endOffset time.Duration) (io.ReadCloser, error) {
	f, err := ioutil.TempFile("", "split*.mp4")
	if err != nil {
		return nil, err
	}

	outName := f.Name()
	log.Print("output will go in ", outName)

	args := []string{
		"-y", // overwrite file
		"-i", v.Filename,
		"-ss", formatDuration(startOffset),
		"-to", formatDuration(endOffset),
		"-async", "1",
		outName,
	}

	log.Println(args)

	stderr, err := util.ExecDebug("ffmpeg", args...)
	if err != nil {
		return nil, fmt.Errorf("%s %v", stderr, err)
	}

	log.Println(stderr)

	return f, nil
}

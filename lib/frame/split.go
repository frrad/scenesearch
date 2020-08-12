package frame

import (
	"fmt"
	"log"
	"time"

	"github.com/frrad/scenesearch/lib/util"
)

func (v *Video) Split(startOffset, endOffset time.Duration, outName string) error {
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
		return fmt.Errorf("%s %v", stderr, err)
	}

	log.Println(stderr)

	return nil
}

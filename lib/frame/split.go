package frame

import (
	"fmt"
	"log"
	"time"

	"github.com/frrad/scenesearch/lib/util"
)

func (v *Video) Split(startOffset, endOffset time.Duration, outName string) error {
	if startOffset > endOffset {
		return fmt.Errorf("start offset (%v) must be <= endOffset (%v)", startOffset, endOffset)
	}

	if endOffset > v.Duration {
		return fmt.Errorf("end offset (%v) must <= end (%v)", endOffset, v.Duration)
	}

	if startOffset < time.Duration(0) {
		return fmt.Errorf("start offset (%v) must >= 0", startOffset)
	}

	fmt.Println(startOffset)
	a, b, e := v.segContaining(startOffset)
	fmt.Println(a, b, e)

	fmt.Println("split")

	return nil
}

func (v Video) segContaining(t time.Duration) (time.Duration, time.Duration, error) {
	i, err := v.segIx(t)
	if err != nil {
		return 0, 0, err
	}

	return v.seg(i)
}

func (v Video) seg(ix int) (time.Duration, time.Duration, error) {
	if ix < 0 || ix > len(v.KeyFrames)-1 {
		return 0, 0, fmt.Errorf("werweasdfasd 52701")
	}

	if ix == len(v.KeyFrames)-1 {
		return v.KeyFrames[ix], v.Duration, nil
	}

	return v.KeyFrames[ix], v.KeyFrames[ix+1], nil
}

// returns the index of the segment containing offset
func (v Video) segIx(offset time.Duration) (int, error) {
	if offset < 0 || offset > v.Duration {
		return 0, fmt.Errorf("asdfadwefe")
	}

	for i, y := range v.KeyFrames {
		if y > offset {
			return i - 1, nil
		}
	}

	return len(v.KeyFrames) - 1, nil
}

func (v *Video) splitReEncode(startOffset, endOffset time.Duration, outName string) error {
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

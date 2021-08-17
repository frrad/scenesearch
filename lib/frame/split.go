package frame

import (
	"fmt"
	"log"
	"time"

	"github.com/frrad/scenesearch/lib/util"
)

type splitPlan struct {
	prefixStart time.Duration
	prefixEnd   time.Duration

	copyStart time.Duration
	copyEnd   time.Duration

	suffixStart time.Duration
	suffixEnd   time.Duration
}

func (v *Video) planSplit(startOffset, endOffset time.Duration) (splitPlan, error) {
	if startOffset > endOffset {
		return splitPlan{}, fmt.Errorf("start offset (%v) must be <= endOffset (%v)", startOffset, endOffset)
	}

	if endOffset > v.Duration {
		return splitPlan{}, fmt.Errorf("end offset (%v) must <= end (%v)", endOffset, v.Duration)
	}

	if startOffset < time.Duration(0) {
		return splitPlan{}, fmt.Errorf("start offset (%v) must >= 0", startOffset)
	}

	plan := splitPlan{}

	// figure out the copy part first
	a, b, err := v.segContaining(startOffset)
	if err != nil {
		return splitPlan{}, err
	}
	plan.copyStart = b
	// if startOffset is on the border, which segment we get back is undefined
	if a == startOffset {
		plan.copyStart = a
	}

	a, b, err = v.segContaining(endOffset)
	plan.copyEnd = a
	if err != nil {
		return splitPlan{}, err
	}
	// again if the offset is on the border, funny things may happen
	if b == endOffset {
		plan.copyEnd = b
	}

	// if cut contains one or fewer segment boundaries
	if plan.copyEnd <= plan.copyStart {
		plan.copyEnd = 0
		plan.copyStart = 0

		plan.prefixStart = startOffset
		plan.prefixEnd = endOffset

		return plan, nil
	}

	if startOffset < plan.copyStart {
		plan.prefixStart = startOffset
		plan.prefixEnd = plan.copyStart
	}

	if plan.copyEnd < endOffset {
		plan.suffixStart = plan.copyEnd
		plan.suffixEnd = endOffset
	}

	return plan, nil
}

func (v *Video) Split(startOffset, endOffset time.Duration, outName string) error {
	sp, err := v.planSplit(startOffset, endOffset)
	if err != nil {
		return err
	}

	log.Printf("split plan %+v", sp)

	var pf, sf string
	if sp.prefixStart < sp.prefixEnd {
		pf, err = v.splitReEncode(sp.prefixStart, sp.prefixEnd)
		if err != nil {
			return err
		}
	}

	if sp.suffixStart < sp.suffixEnd {
		sf, err = v.splitReEncode(sp.suffixStart, sp.suffixEnd)
		if err != nil {
			return err
		}
	}

	fmt.Println(pf, sf)

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

func (v *Video) splitReEncode(startOffset, endOffset time.Duration) (string, error) {
	outName := fmt.Sprintf("%s-%d-%d.mp4", v.Filename, startOffset, endOffset)

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
		return "", fmt.Errorf("%s %v", stderr, err)
	}

	log.Println(stderr)

	return outName, nil
}

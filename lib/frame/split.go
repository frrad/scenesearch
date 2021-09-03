package frame

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/frrad/scenesearch/lib/util"
)

const cacheName = "cache"

func (v *Video) splitDoneFileName(start, end time.Duration) string {
	return fmt.Sprintf("%s/%s-%d-%d.mp4", cacheName, v.HashString[:7], start, end)
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func tempFileName(suffix string) (string, error) {
	now := time.Now().UnixNano()
	randomStr, err := randomHex(4)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%d-%s%s", cacheName, now, randomStr, suffix), nil
}

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

// Split splits
//
// https://stackoverflow.com/a/63604858
func (v *Video) Split(startOffset, endOffset time.Duration) (string, error) {
	completeSplitName := v.splitDoneFileName(startOffset, endOffset)

	if _, err := os.Stat(completeSplitName); !os.IsNotExist(err) {
		log.Println(completeSplitName, "already exists, not recreating")
		return completeSplitName, nil
	}

	sp, err := v.planSplit(startOffset, endOffset)
	if err != nil {
		return "", err
	}

	log.Printf("split plan %+v", sp)

	var pf, sf string
	if sp.prefixStart < sp.prefixEnd {
		pf, err = v.splitReEncode(sp.prefixStart, sp.prefixEnd)
		if err != nil {
			return "", err
		}
	}

	if sp.suffixStart < sp.suffixEnd {
		sf, err = v.splitReEncode(sp.suffixStart, sp.suffixEnd)
		if err != nil {
			return "", err
		}
	}

	concatInput := ""
	if pf != "" {
		concatInput += fmt.Sprintf("file '%s'\n", pf)
	}
	if sp.copyStart < sp.copyEnd {
		concatInput += fmt.Sprintf("file '%s'\n", v.Filename)
		concatInput += fmt.Sprintf("inpoint %f\n", sp.copyStart.Seconds())
		concatInput += fmt.Sprintf("outpoint %f\n", sp.copyEnd.Seconds())
	}
	if sf != "" {
		concatInput += fmt.Sprintf("file '%s'\n", sf)
	}

	concatFileName := fmt.Sprintf("concatinstructions-%d-%d.txt", startOffset, endOffset)
	ioutil.WriteFile(concatFileName, []byte(concatInput), 0744)

	tfn, err := tempFileName(".mp4")
	if err != nil {
		return "", err
	}

	args := []string{
		"-f", "concat",
		"-i", concatFileName,
		"-c", "copy",
		tfn,
	}

	log.Println(args)

	ffmpegOutput, err := util.ExecDebug("ffmpeg", args...)
	if err != nil {
		return "", fmt.Errorf("%s %v", ffmpegOutput, err)
	}

	cmd := exec.Command("mv", tfn, completeSplitName)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	log.Println(ffmpegOutput)

	cmd = exec.Command("rm", concatFileName)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return completeSplitName, nil
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
	outName := v.splitDoneFileName(startOffset, endOffset)

	if _, err := os.Stat(outName); !os.IsNotExist(err) {
		log.Println(outName, "already exists, not recreating")
		return outName, nil
	}

	tfn, err := tempFileName(".mp4")
	if err != nil {
		return "", err
	}

	args := []string{
		"-y", // overwrite file
		"-i", v.Filename,
		"-ss", formatDuration(startOffset),
		"-to", formatDuration(endOffset),
		"-async", "1",
		"-profile:v", v.Profile,
		tfn,
	}

	log.Println(args)

	stderr, err := util.ExecDebug("ffmpeg", args...)
	if err != nil {
		return "", fmt.Errorf("%s %v", stderr, err)
	}

	log.Println(stderr)

	cmd := exec.Command("mv", tfn, outName)
	_, err = cmd.CombinedOutput()

	return outName, err
}

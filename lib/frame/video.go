package frame

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Video struct {
	Filename  string
	Duration  time.Duration
	Profile   string
	KeyFrames []time.Duration
}

func NewVideo(filename string) (Video, error) {
	v := Video{
		Filename: filename,
	}

	frames, err := v.keyFrames()
	if err != nil {
		return Video{}, err
	}
	v.KeyFrames = frames

	err = v.populateDuration()
	if err != nil {
		return Video{}, err
	}

	err = v.populateProfile()
	if err != nil {
		return Video{}, err
	}

	return v, nil
}

func (v *Video) populateDuration() error {
	n := "ffprobe"
	cmd := []string{
		"-v",
		"error",
		"-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", v.Filename}

	ret, err := exec.Command(n, cmd...).Output()
	if err != nil {
		return fmt.Errorf("error executing %s: %w", n, err)
	}

	returnedString := strings.TrimSuffix(string(ret), "\n")
	dur, err := parseStringAsDurationSec(returnedString)
	if err != nil {
		return err
	}

	v.Duration = dur

	return nil
}

func (v *Video) populateProfile() error {
	n := "ffprobe"
	cmd := []string{
		"-v", "error", "-select_streams", "v", "-show_entries", "stream=profile", "-of", "csv=p=0",
		v.Filename}

	ret, err := exec.Command(n, cmd...).Output()
	if err != nil {
		return fmt.Errorf("error executing %s: %w", n, err)
	}

	returnedString := strings.TrimSuffix(string(ret), "\n")

	v.Profile = returnedString

	return nil
}

func (v Video) keyFrames() ([]time.Duration, error) {
	n := "ffprobe"
	cmd := []string{"-v", "error", "-select_streams", "v:0", "-skip_frame", "nokey", "-show_entries", "frame=pkt_pts_time", "-of", "csv=p=0", v.Filename}

	ret, err := exec.Command(n, cmd...).Output()
	if err != nil {
		return []time.Duration{}, fmt.Errorf("error executing %s: %w", n, err)
	}

	s := strings.Split(string(ret), "\n")

	ans := []time.Duration{}
	for _, x := range s {
		if x == "" {
			continue
		}

		z, err := parseStringAsDurationSec(x)
		if err != nil {
			return []time.Duration{}, err
		}
		ans = append(ans, z)
	}

	return ans, nil
}

func parseStringAsDurationSec(x string) (time.Duration, error) {
	z, err := strconv.ParseFloat(x, 64)
	if err != nil {
		return time.Duration(0), fmt.Errorf("error parsing as second: %w", err)
	}
	return time.Duration(float64(time.Second) * z), nil
}

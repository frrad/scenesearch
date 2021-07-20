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

	return v, nil
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

		z, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return []time.Duration{}, fmt.Errorf("error executing %s: %w", n, err)
		}
		ans = append(ans, time.Second*time.Duration(z))

		fmt.Println("asdf", x)
	}

	return ans, nil
}

package frame

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/frrad/scenesearch/lib/util"
)

func (v *Video) Length() (time.Duration, error) {
	log.Println("getting duration")

	args := []string{"-i", v.Filename}
	ans, err := util.ExecDebug("ffprobe", args...)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(ans, "\n")
	for _, l := range lines {
		ans, ok := parseDuration(l)
		if ok {
			return ans, nil
		}
	}

	return 0, fmt.Errorf("couldn't find duration in output")
}

func parseDuration(l string) (time.Duration, bool) {
	d := "Duration:"

	l = strings.Trim(l, " ")

	if len(l) < len(d) {
		return 0, false
	}

	if l[:len(d)] != d {
		return 0, false
	}

	parts := strings.Split(l, ",")
	if len(parts) < 1 {
		return 0, false
	}

	dStr := parts[0]

	tParts := strings.Split(dStr, ":")

	// first part is the text "duration" so discard
	tParts = tParts[1:]

	if len(tParts) != 3 {
		log.Fatalf("expected 3 parts of %v", tParts)
	}

	hStr, mStr, sStr := tParts[0], tParts[1], tParts[2]
	h, err := strconv.ParseUint(strings.Trim(hStr, " "), 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	m, err := strconv.ParseUint(mStr, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	s, err := strconv.ParseFloat(sStr, 64)
	if err != nil {
		log.Fatal(err)
	}

	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second, true
}

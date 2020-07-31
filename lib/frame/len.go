package frame

import (
	"log"
	"time"

	"github.com/frrad/scenesearch/lib/util"
)

func (v *Video) Length() (time.Duration, error) {
	log.Println("getting duration")

	args := []string{"-i", v.Filename}
	ans, err := util.ExecDebug("ffmpeg", args...)
	if err != nil {
		return 0, err
	}

	log.Println(ans)

	return 0, nil
}

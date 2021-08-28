package main

import (
	"log"
	"sync"

	"github.com/frrad/scenesearch/lib/frame"
)

type VideoFrames struct {
	*sync.RWMutex
	State map[string]*frame.Video
}

var VideoFrameCache VideoFrames

func init() {
	log.Println("initializing...")

	VideoFrameCache = VideoFrames{
		RWMutex: &sync.RWMutex{},
		State:   map[string]*frame.Video{},
	}
}

func (v *VideoFrames) GetFrame(name string) (*frame.Video, error) {
	v.RLock()
	ans, ok := v.State[name]
	v.RUnlock()

	if ok {
		return ans, nil
	}

	v.Lock()
	defer v.Unlock()

	newVideo, err := frame.NewVideo(name)
	if err != nil {
		return nil, err
	}

	v.State[name] = &newVideo
	return &newVideo, nil
}

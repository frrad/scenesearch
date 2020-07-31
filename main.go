package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/frrad/scenesearch/lib/frame"
)

const (
	frameRoute = "/frame"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc(frameRoute, handleFrame)
	http.HandleFunc("/compare", handleCompare)
	http.HandleFunc("/done", handleDone)

	port := ":8080"
	log.Println("serving on port", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

type frameReq struct {
	Offset int64 // milliseconds
	file   string
}

func (f frameReq) String() string {
	return fmt.Sprintf("%s?offset=%d&file=%s", frameRoute, f.Offset, f.file)
}

func handleFrame(w http.ResponseWriter, r *http.Request) {
	offsets := r.URL.Query()["offset"]
	if len(offsets) != 1 {
		w.WriteHeader(http.StatusBadRequest)
	}

	offsetUint, err := strconv.ParseUint(offsets[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	files := r.URL.Query()["file"]
	if len(files) != 1 {
		w.WriteHeader(http.StatusBadRequest)
	}

	v := frame.Video{
		Filename: files[0],
	}

	frame, err := v.ExtractFrame(time.Duration(offsetUint) * time.Millisecond)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	_, err = io.Copy(w, frame)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	err = frame.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
}

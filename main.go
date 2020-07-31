package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/frrad/scenesearch/lib/frame"
)

func main() {
	http.HandleFunc("/frame", handleFrame)
	http.HandleFunc("/compare", handleCompare)

	log.Fatal(http.ListenAndServe(":8080", nil))
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

	v := frame.Video{
		Filename: "input.mp4",
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

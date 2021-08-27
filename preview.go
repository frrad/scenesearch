package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/frrad/scenesearch/lib/frame"
)

type previewReq struct {
	File  string
	Start int64
	End   int64
}

func (p previewReq) AsLink(text string) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"%s\"> %s </a>", p.String(), text))
}

func (p previewReq) String() string {
	return fmt.Sprintf("%s?start=%d&end=%d&file=%s", previewRoute, p.Start, p.End, p.File)
}

func (p previewReq) Preview() (string, error) {
	v, err := frame.NewVideo(p.File)
	if err != nil {
		return "", err
	}

	return v.Split(time.Duration(p.Start)*time.Millisecond, time.Duration(p.End)*time.Millisecond)
}

func numFromURL(url *url.URL, param string) (uint64, error) {
	offsets := url.Query()[param]
	if len(offsets) != 1 {
		return 0, fmt.Errorf("expected exactly 1 value for param %s", param)
	}

	offsetUint, err := strconv.ParseUint(offsets[0], 10, 64)
	if err != nil {
		return 0, err
	}

	return offsetUint, nil
}

func handlePreview(w http.ResponseWriter, r *http.Request) {
	log.Println("got preview req", r)

	req := previewReq{}

	start, err := numFromURL(r.URL, "start")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	req.Start = int64(start)

	end, err := numFromURL(r.URL, "end")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	req.End = int64(end)

	files := r.URL.Query()["file"]
	if len(files) != 1 {
		w.WriteHeader(http.StatusBadRequest)
	}
	req.File = files[0]

	log.Println("parsed req", req)

	splitFilename, err := req.Preview()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, splitFilename)
}

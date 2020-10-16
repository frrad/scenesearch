package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/frrad/scenesearch/lib/frame"
)

type frameReq struct {
	Offset int64 // milliseconds
	file   string
	Width  int
}

type frameRange struct {
	StartOffset int64 // ms
	EndOffset   int64 // ms
	Shots       int   // at least 2
	Width       int

	File string
}

const rangeHTML = `	<table>
<tr>
    {{ range $index, $value := .Frames }}
    <td>{{$value.Offset}}</td>
    {{ end }}
</tr>

<tr>
    {{ range $index, $value := .Frames }}
    <td><img src="{{$value}}" width="{{.Width}}px"></td>
    {{ end }}
</tr>
</table>
`

var rangeTemplate = template.Must(template.New("").Parse(rangeHTML))

func (r frameRange) Frames() []frameReq {
	ans := make([]frameReq, r.Shots)

	interval := (r.EndOffset - r.StartOffset) / (int64(r.Shots) - 1)
	o := r.StartOffset

	for i := 0; i < r.Shots; i++ {
		ans[i] = frameReq{
			Offset: o,
			file:   r.File,
			Width:  r.Width,
		}
		o += interval
	}

	// fix possible rounding issues for last frame offset
	ans[r.Shots-1] = frameReq{
		Offset: r.EndOffset,
		file:   r.File,
		Width:  r.Width,
	}

	return ans
}

func (r frameRange) Table() template.HTML {
	b := bytes.Buffer{}

	err := rangeTemplate.Execute(&b, r)
	if err != nil {
		fmt.Fprintf(&b, "%+v", err)
	}

	return template.HTML(b.String())
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

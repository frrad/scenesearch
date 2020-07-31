package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/frrad/scenesearch/lib/frame"
)

const compareHtml = `
<html>

<table>
<tr>
<td><img src="/frame?offset={{.Offset1}}" width="{{.Width}}px"></td>
<td><img src="/frame?offset={{.Offset2}}" width="{{.Width}}px"></td>
</tr>
</table>

{{.State}}

</html>
`

var compareTemplate = template.Must(template.New("").Parse(compareHtml))

type ComparePageData struct {
	Offset1 uint64
	Offset2 uint64
	Width   uint64

	State string
}

type SearchState struct {
	FileName string
	Length   time.Duration

	Cuts     []time.Duration
	Segments []Segment
}

type Segment struct {
	Start time.Duration
	End   time.Duration
}

var initialState = SearchState{
	FileName: "input.mp4",
}

func handleCompare(w http.ResponseWriter, r *http.Request) {
	log.Println("got compare request")
	stateStrs := r.URL.Query()["state"]
	if len(stateStrs) != 1 {
		log.Println("not one state value")
		initialState.Reset(w, r)
		return
	}

	state := &SearchState{}
	err := state.Decode(stateStrs[0])
	if err != nil {
		log.Printf("error decoding: %s", err)
		initialState.Reset(w, r)
		return
	}

	changed, err := state.Normalize()
	if err != nil {
		log.Printf("error normalizing: %s", err)
		initialState.Reset(w, r)
		return
	}

	if changed {
		state.Reset(w, r)
		return
	}

	if state.FileName == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = compareTemplate.Execute(w, ComparePageData{
		Offset1: 0,
		Offset2: 1000,
		Width:   500,

		State: fmt.Sprintf("%v", state),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *SearchState) Encode() (string, error) {
	b := bytes.Buffer{}

	// Create an encoder and send a value.
	enc := gob.NewEncoder(&b)
	err := enc.Encode(s)
	if err != nil {
		return "", err
	}

	encodedStr := base64.URLEncoding.EncodeToString(b.Bytes())
	log.Printf("encoded: %s", encodedStr)
	return encodedStr, nil
}

func (s *SearchState) Decode(in string) error {
	log.Printf("decoding: %s", in)

	b, err := base64.URLEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	enc := gob.NewDecoder(bytes.NewBuffer(b))
	err = enc.Decode(s)
	if err != nil {
		return err
	}

	return nil
}

func (s *SearchState) Reset(w http.ResponseWriter, r *http.Request) {
	stateStr, err := s.Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/compare?state="+stateStr, http.StatusSeeOther)
}

func (s *SearchState) Normalize() (bool, error) {
	changed := false

	vid := frame.Video{
		Filename: s.FileName,
	}

	if s.Length == 0 {
		dur, err := vid.Length()
		if err != nil {
			log.Printf("err getting len %s", err)
			return changed, err
		}
		s.Length = dur
		s.Length = 5 * time.Minute // hack for now
		changed = true
	}

	if len(s.Cuts) == 0 && len(s.Segments) == 0 {
		s.Cuts = []time.Duration{0, s.Length}
		changed = true
	}

	return changed, nil
}

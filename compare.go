package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/frrad/scenesearch/lib/frame"
)

const compareHtml = `
<html>

Gap Size: {{.GapSize}}

<table>
<tr>
<td>{{.Offset1.Milliseconds}}</td>
<td>{{.Offset2.Milliseconds}}</td>
</tr>
<tr>
<td><img src="/frame?offset={{.Offset1.Milliseconds}}" width="{{.Width}}px"></td>
<td><img src="/frame?offset={{.Offset2.Milliseconds}}" width="{{.Width}}px"></td>
</tr>
</table>

<h3>
<a href="/compare?state={{.IfSame}}"> Same </a>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
<a href="/compare?state={{.IfDiff}}"> Diff </a>
</h3>

<pre><code>
{{.State}}
</code></pre>

</html>
`

var compareTemplate = template.Must(template.New("").Parse(compareHtml))

type ComparePageData struct {
	GapSize time.Duration
	Offset1 time.Duration
	Offset2 time.Duration
	Width   uint64

	IfSame string
	IfDiff string

	State string
}

type SearchState struct {
	FileName string
	Length   time.Duration

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

	err = state.Normalize()
	if err != nil {
		log.Printf("error normalizing: %s", err)
		initialState.Reset(w, r)
		return
	}

	a, b, err := state.MaxGap()

	stateJson, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	same, err := state.IfSame(a, b).Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	diff, err := state.IfDifferent(a, b).Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	err = compareTemplate.Execute(w, ComparePageData{
		GapSize: b - a,
		Offset1: a,
		Offset2: b,
		Width:   500,

		IfSame: same,
		IfDiff: diff,

		State: string(stateJson),
	})
	if err != nil {
		fmt.Fprintf(w, "%+v", err)
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
	log.Printf("decoding: ...")

	b, err := base64.URLEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	enc := gob.NewDecoder(bytes.NewBuffer(b))
	err = enc.Decode(s)
	if err != nil {
		return err
	}

	log.Printf("got:%+v", s)

	return nil
}

func (s *SearchState) Reset(w http.ResponseWriter, r *http.Request) {
	stateStr, err := s.Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/compare?state="+stateStr, http.StatusSeeOther)
}

func (s *SearchState) MaxGap() (time.Duration, time.Duration, error) {
	x := []Segment{}
	for _, seg := range s.Segments {
		x = append(x, seg)
	}

	sort.Slice(x, func(i, j int) bool {
		return x[i].Start < x[j].Start
	})

	max, a, b := time.Duration(0), time.Duration(0), time.Duration(0)
	for i := 0; i < len(x)-1; i++ {
		gap := x[i+1].Start - x[i].End
		if gap > max {
			max = gap
			a, b = x[i].End, x[i+1].Start
		}
		if gap < 0 {
			return 0, 0, fmt.Errorf("negative gap between %v and %v", x[i], x[i+1])
		}
	}
	return a, b, nil
}

func (s *SearchState) IfDifferent(a, b time.Duration) *SearchState {
	t := s.Copy()
	new := (a + b) / 2
	t.Segments = append(t.Segments, Segment{Start: new, End: new})
	t.SortSegs()
	return &t
}

func (s *SearchState) IfSame(a, b time.Duration) *SearchState {
	t := s.Copy()
	t.Segments = append(t.Segments, Segment{Start: a, End: b})

	t.Meld()

	return &t
}

func (s *SearchState) Meld() {
	s.SortSegs()
	if len(s.Segments) <= 1 {
		return
	}

	a, b := 0, 1
	for b < len(s.Segments) {
		s.Segments[a+1] = s.Segments[b]

		if s.Segments[a].End == s.Segments[a+1].Start {
			s.Segments[a].End = s.Segments[a+1].End
			b++
			continue
		}

		a++
		b++
	}

	s.Segments = s.Segments[:a+1]
}

func (s *SearchState) SortSegs() {
	sort.Slice(s.Segments, func(i, j int) bool {
		return s.Segments[i].Start < s.Segments[j].Start
	})
}

func (s *SearchState) Copy() SearchState {
	segCopy := []Segment{}
	for _, x := range s.Segments {
		segCopy = append(segCopy, x)
	}

	return SearchState{
		FileName: s.FileName,
		Length:   s.Length,
		Segments: segCopy,
	}
}

func (s *SearchState) Normalize() error {
	if s.FileName == "" {
		return fmt.Errorf("need filename")
	}

	vid := frame.Video{
		Filename: s.FileName,
	}

	if s.Length == 0 {
		dur, err := vid.Length()
		if err != nil {
			log.Printf("err getting len %s", err)
			return err
		}
		s.Length = dur
		s.Length = 5 * time.Minute // hack for now
	}

	if len(s.Segments) < 2 {
		s.Segments = []Segment{
			{0, 0},
			{Start: s.Length, End: s.Length},
		}
	}

	if s.Segments == nil {
		s.Segments = []Segment{}
	}

	return nil
}

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
)

type SearchState struct {
	FileName string
	Length   time.Duration

	Segments    []Segment
	Breakpoints map[time.Duration]struct{}
}

type Segment struct {
	Start time.Duration
	End   time.Duration
}

var initialState = SearchState{
	FileName: "input.mp4",
}

func (seg *Segment) Len() time.Duration {
	return seg.End - seg.Start
}

func (seg *Segment) Frame(f string, pct float64) frameReq {

	x := seg.Start + time.Duration(pct*float64(seg.End-seg.Start))

	return frameReq{
		Offset: x.Milliseconds(),
		file:   f,
	}
}

func (state *SearchState) Done() bool {
	_, _, err := state.MaxGap()
	if err == ErrDone {
		return true
	}
	return false
}

func (s *SearchState) Encode() string {
	b := bytes.Buffer{}

	// Create an encoder and send a value.
	enc := gob.NewEncoder(&b)
	err := enc.Encode(s)
	if err != nil {
		log.Fatal("never happens")
	}

	encodedStr := base64.URLEncoding.EncodeToString(b.Bytes())
	return encodedStr
}

func (s *SearchState) PercentSegmented() float64 {
	total := time.Duration(0)
	for _, x := range s.Segments {
		total += x.End - x.Start
	}

	return 100. * float64(total) / float64(s.Length)
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

	return nil
}

func (s *SearchState) ComparisonPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.AsCompareLink(), http.StatusSeeOther)
}

func (s *SearchState) DonePage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/done?state="+s.Encode(), http.StatusSeeOther)
}

func (s *SearchState) AsCompareLink() string {
	return "/compare?state=" + s.Encode()
}

var ErrDone = errors.New("no more gaps")

const quantum = time.Millisecond

func (s *SearchState) MaxGap() (time.Duration, time.Duration, error) {
	x := []Segment{}
	for _, seg := range s.Segments {
		x = append(x, seg)
	}

	sort.Slice(x, func(i, j int) bool {
		return x[i].Start < x[j].Start
	})

	max, a, b := time.Duration(0), time.Duration(0), time.Duration(0)
	found := false
	for i := 0; i < len(x)-1; i++ {
		f, g := x[i].End, x[i+1].Start

		gap := g - f
		if gap < 0 {
			return 0, 0, fmt.Errorf("negative gap between %v and %v", x[i], x[i+1])
		}

		if gap > max {
			_, ok1 := s.Breakpoints[f]
			_, ok2 := s.Breakpoints[g]
			if (ok1 || ok2) && gap <= quantum {
				continue
			}

			found = true
			max = gap
			a, b = f, g
		}
	}

	if !found {
		return 0, 0, ErrDone
	}

	return a, b, nil
}

func (s *SearchState) IfDifferent(a, b time.Duration) *SearchState {
	t := s.Copy()
	new := (a + b) / 2
	new = new.Round(quantum)

	if new != a && new != b {
		t.Segments = append(t.Segments, Segment{Start: new, End: new})
		t.SortSegs()
		return &t
	}

	t.Breakpoints[a] = struct{}{}
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
		return s.Segments[i].Start < s.Segments[j].Start || s.Segments[i].Start == s.Segments[j].Start && s.Segments[i].End < s.Segments[j].End
	})
}

func (s *SearchState) Copy() SearchState {
	segCopy := []Segment{}
	for _, x := range s.Segments {
		segCopy = append(segCopy, x)
	}

	bCop := map[time.Duration]struct{}{}
	for x := range s.Breakpoints {
		bCop[x] = struct{}{}
	}

	return SearchState{
		FileName:    s.FileName,
		Length:      s.Length,
		Segments:    segCopy,
		Breakpoints: bCop,
	}
}

func (s *SearchState) JSON() string {
	x, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Fatal("never happens")
	}

	return string(x)
}

func (s *SearchState) Normalize() error {
	if s.FileName == "" {
		return fmt.Errorf("need filename")
	}

	vid, err := VideoFrameCache.GetFrame(s.FileName)
	if err != nil {
		log.Printf("err getting len %s", err)
		return err
	}

	if s.Length == 0 {
		s.Length = vid.Duration
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

	if s.Breakpoints == nil {
		s.Breakpoints = map[time.Duration]struct{}{}
	}

	return nil
}

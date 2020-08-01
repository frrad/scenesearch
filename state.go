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

	"github.com/frrad/scenesearch/lib/frame"
)

type SearchState struct {
	FileName string
	Length   time.Duration

	Segments    []Segment
	Breakpoints map[time.Duration]struct{}
	Done        bool
}

type Segment struct {
	Start time.Duration
	End   time.Duration
}

var initialState = SearchState{
	FileName: "input.mp4",
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
	log.Printf("encoded: %s", encodedStr)
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

	log.Printf("got:%+v", s)

	return nil
}

func (s *SearchState) ComparisonPage(w http.ResponseWriter, r *http.Request) {
	stateStr := s.Encode()

	http.Redirect(w, r, "/compare?state="+stateStr, http.StatusSeeOther)
}

func (s *SearchState) DonePage(w http.ResponseWriter, r *http.Request) {
	stateStr := s.Encode()

	http.Redirect(w, r, "/done?state="+stateStr, http.StatusSeeOther)
}

var ErrDone = errors.New("no more gaps")

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
		if _, ok := s.Breakpoints[x[i+1].Start]; ok {
			continue
		}
		if _, ok := s.Breakpoints[x[i].End]; ok {
			continue
		}
		gap := x[i+1].Start - x[i].End
		if gap > max {
			found = true
			max = gap
			a, b = x[i].End, x[i+1].Start
		}
		if gap < 0 {
			return 0, 0, fmt.Errorf("negative gap between %v and %v", x[i], x[i+1])
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
	new = new.Round(time.Millisecond)

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

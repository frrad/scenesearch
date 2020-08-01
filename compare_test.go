package main

import (
	"bytes"
	"testing"
)

func TestCompare(t *testing.T) {
	w := &bytes.Buffer{}

	err := compareTemplate.Execute(w, ComparePageData{
		GapSize: 1,
		Frame1: frameReq{
			Offset: 10,
			file:   "jim",
		},
		Frame2: frameReq{
			Offset: 11,
			file:   "bob",
		},
		Width: 500,

		IfSame: &SearchState{},
		IfDiff: &SearchState{},

		SS: &SearchState{
			Segments: []Segment{{Start: 0, End: 1}},
			Length:   10,
		},
	})

	if err != nil {
		t.Errorf("%s", err)
	}
}

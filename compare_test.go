package main

import (
	"bytes"
	"testing"
)

func TestCompare(t *testing.T) {
	w := &bytes.Buffer{}

	err := compareTemplate.Execute(w, ComparePageData{
		GapSize: 1,
		Range: frameRange{
			StartOffset: 10,
			EndOffset:   11,
			Shots:       5,
			Width:       200,
			File:        "jimbob",
		},

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

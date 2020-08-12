package main

import (
	"bytes"
	"testing"
)

func TestDone(t *testing.T) {
	w := &bytes.Buffer{}

	err := doneTemplate.Execute(w, DonePageData{
		State: &SearchState{},
	})

	if err != nil {
		t.Errorf("%s", err)
	}
}

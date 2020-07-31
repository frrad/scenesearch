package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"html/template"
	"net/http"
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

type SearchState struct{}

func handleCompare(w http.ResponseWriter, r *http.Request) {
	stateStrs := r.URL.Query()["state"]
	if len(stateStrs) != 1 {
		initialState := SearchState{}
		stateStr, err := initialState.Encode()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		http.Redirect(w, r, "/compare?state="+stateStr, 301)
		return
	}

	state := &SearchState{}
	err := state.Decode(stateStrs[0])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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

	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func (s *SearchState) Decode(in string) error {
	b, err := base64.StdEncoding.DecodeString(in)
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

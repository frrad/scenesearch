package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

const compareHtml = `
<html>

<head>

<script>
document.addEventListener("keypress", function(event) {
if (event.keyCode == 121) {
location.href = "/compare?state={{.IfSame}}";
}
if (event.keyCode == 110) {
location.href = "/compare?state={{.IfDiff}}";
}


});
</script>

</head>

{{if .SS.Done}}<h1> DONE! </h1> {{end}}

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

	SS    *SearchState
	State string
}

func handleCompare(w http.ResponseWriter, r *http.Request) {
	log.Println("got compare request")
	stateStrs := r.URL.Query()["state"]
	if len(stateStrs) != 1 {
		log.Println("not one state value")
		initialState.ComparisonPage(w, r)
		return
	}

	state := &SearchState{}
	err := state.Decode(stateStrs[0])
	if err != nil {
		log.Printf("error decoding: %s", err)
		initialState.ComparisonPage(w, r)
		return
	}

	err = state.Normalize()
	if err != nil {
		log.Printf("error normalizing: %s", err)
		initialState.ComparisonPage(w, r)
		return
	}

	a, b, err := state.MaxGap()
	if err == ErrDone {
		state.Done = true
		state.DonePage(w, r)
	}

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

		SS:    state,
		State: string(stateJson),
	})
	if err != nil {
		fmt.Fprintf(w, "%+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

package main

import (
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
    location.href = "/compare?state={{.IfSame.Encode}}";
  }
  if (event.keyCode == 110) {
    location.href = "/compare?state={{.IfDiff.Encode}}";
  }
});
</script>

</head>

<b>Gap Size:</b> {{.GapSize}}                                       &nbsp;&nbsp;
<b>Percent Segmented:</b> {{printf "%.2f" .SS.PercentSegmented}} %  &nbsp;&nbsp;
<b>Num  Segments:</b> {{ len .SS.Segments}}


{{.Range.Table}}

<h3>
<a href="/compare?state={{.IfSame.Encode}}"> Same </a>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
<a href="/compare?state={{.IfDiff.Encode}}"> Diff </a>
</h3>

<pre><code>
{{.SS.JSON}}
</code></pre>

</html>
`

var compareTemplate = template.Must(template.New("").Parse(compareHtml))

type ComparePageData struct {
	GapSize time.Duration
	Range   frameRange

	IfSame *SearchState
	IfDiff *SearchState

	SS *SearchState
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
		state.DonePage(w, r)
		return
	}

	err = compareTemplate.Execute(w, ComparePageData{
		GapSize: b - a,
		Range: frameRange{
			StartOffset: a.Milliseconds(),
			EndOffset:   b.Milliseconds(),
			Shots:       5,
			Width:       200,
			File:        state.FileName,
		},

		IfSame: state.IfSame(a, b),
		IfDiff: state.IfDifferent(a, b),

		SS: state,
	})
	if err != nil {
		fmt.Fprintf(w, "%+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

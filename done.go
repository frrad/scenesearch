package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const doneHtml = `
<html>

<head>
</head>

<pre><code>
{{.State}}
</code></pre>

</html>
`

var doneTemplate = template.Must(template.New("").Parse(doneHtml))

type DonePageData struct {
	State string
}

func handleDone(w http.ResponseWriter, r *http.Request) {
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

	stateJson, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	err = doneTemplate.Execute(w, DonePageData{
		State: string(stateJson),
	})
	if err != nil {
		fmt.Fprintf(w, "%+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

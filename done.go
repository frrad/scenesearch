package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const doneHtml = `
<html>

<head>
</head>

<form action="{{.SplitRoute}}" method="post">
    <button name="state" value="{{.State.Encode}}">Split!</button>
</form>

<table>

<tr>
<th> Index </th>
<th> Length </th>
<th> Start </th>
<th> End </th>
</tr>

{{range $index, $element := .State.Segments}}
<tr>
<td> {{$index}} </td>
<td> {{$element.Len}} </td>
<td> {{$element.Start}} </td>
<td> {{$element.End}} </td>
</tr>
{{end}}
</table>

</html>
`

var doneTemplate = template.Must(template.New("").Parse(doneHtml))

type DonePageData struct {
	State      *SearchState
	SplitRoute string
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

	err = doneTemplate.Execute(w, DonePageData{
		SplitRoute: splitRoute,
		State:      state,
	})
	if err != nil {
		fmt.Fprintf(w, "%+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

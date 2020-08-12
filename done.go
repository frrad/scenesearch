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
<style>
table, th, td {
  border: 1px solid black;
  text-align: left;
  vertical-align: top;
}
</style>
</head>

<form action="{{.SplitRoute}}" method="post">
    <button name="state" value="{{.State.Encode}}">Split!</button>
</form>

<table>

<tr>
<th> Index </th>
<th> Length </th>
</tr>

{{range $index, $element := .State.Segments}}
<tr>
<td> {{$index}} </td>
<td> {{$element.Len}} </td>
<td> <img src="{{$element.Frame "input.mp4" 0.1}}" width="250px"> </td>
<td> <img src="{{$element.Frame "input.mp4" 0.5}}" width="250px"> </td>
<td> <img src="{{$element.Frame "input.mp4" 0.9}}" width="250px"> </td>
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

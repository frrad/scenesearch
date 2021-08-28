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
  {{ range $index, $row := .Table}}
  <tr>
	{{ range $index, $col := $row }}
	<td>
	  {{$col}}
	</td>
	{{end}}
  </tr>
  {{end}}
</table>

</html>
`

var doneTemplate = template.Must(template.New("").Parse(doneHtml))

type DonePageData struct {
	State      *SearchState
	SplitRoute string
	Table      [][]template.HTML
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
		Table:      buildTable(state),
	})
	if err != nil {
		fmt.Fprintf(w, "%+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func buildTable(x *SearchState) [][]template.HTML {
	ans := [][]template.HTML{{"#", "duration", "", "preview"}}

	for i, seg := range x.Segments {
		row := []template.HTML{
			template.HTML(fmt.Sprintf("%d", i)),
			template.HTML(fmt.Sprintf("%s", seg.Len())),
			frameRange{
				StartOffset: seg.Start.Milliseconds(),
				EndOffset:   seg.End.Milliseconds(),
				Shots:       5,
				Width:       200,
				File:        x.FileName,
			}.Table(),
			previewReq{
				File:  x.FileName,
				Start: seg.Start.Milliseconds(),
				End:   seg.End.Milliseconds(),
			}.AsLink("Preview"),
			template.HTML(fmt.Sprintf("<a href=\"%s\">Bust</a>", x.Bust(i).AsCompareLink())),
		}
		ans = append(ans, row)
	}

	return ans
}

func (s *SearchState) Bust(i int) *SearchState {
	t := s.Copy()

	a := Segment{
		Start: t.Segments[i].Start,
		End:   t.Segments[i].Start,
	}
	b := Segment{
		Start: t.Segments[i].End,
		End:   t.Segments[i].End,
	}

	x := make([]Segment, len(t.Segments)+1)
	copy(x[:i], t.Segments[:i])
	x[i] = a
	x[i+1] = b
	copy(x[i+2:], t.Segments[i+1:])

	t.Segments = x

	return &t
}

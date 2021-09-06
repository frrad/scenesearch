package main

import (
	"html/template"
	"log"
	"net/http"
)

const startHTML = `
<html>

<form action="{{.StartRoute}}">
  <input type="text" id="filename" name="filename">
  <input type="submit">
</form>

</html>
`

type StartPageData struct {
	StartRoute string
}

var startTemplate = template.Must(template.New("").Parse(startHTML))

func handleStart(w http.ResponseWriter, r *http.Request) {
	log.Println(r)

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	ans, ok := r.Form["filename"]
	if !ok {
		startTemplate.Execute(w, StartPageData{
			StartRoute: startRoute,
		})
		return
	}

	if len(ans) != 1 {
		log.Fatal("not one state")
	}

	state := &SearchState{
		FileName: ans[0],
	}

	err = state.Normalize()
	if err != nil {
		log.Printf("error normalizing: %s", err)
		initialState.ComparisonPage(w, r)
		return
	}

	http.Redirect(w, r, state.AsCompareLink(), http.StatusSeeOther)
}

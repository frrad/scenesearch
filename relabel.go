package main

import (
	"log"
	"net/http"
	"strconv"
)

func handleReLabel(w http.ResponseWriter, r *http.Request) {
	log.Println(r)

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	ans, ok := r.Form["state"]
	if !ok {
		log.Fatal("no state!")
	}

	if len(ans) != 1 {
		log.Fatal("not one state")
	}

	state := &SearchState{}
	err = state.Decode(ans[0])
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

	for k, v := range r.Form {
		if k == "state" {
			continue
		}

		if len(v) == 0 {
			continue
		}

		if len(v[0]) == 0 {
			continue
		}

		x, err := strconv.Atoi(k)
		if err != nil {
			log.Fatal(err)
		}

		state.Segments[x].Label = v[0]
	}

	http.Redirect(w, r, "/done?state="+state.Encode(), http.StatusSeeOther)
}

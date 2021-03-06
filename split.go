package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/frrad/scenesearch/lib/frame"
)

func handleSplit(w http.ResponseWriter, r *http.Request) {
	log.Println("got split request!")

	r.ParseForm()
	states := r.Form["state"]
	log.Println(r.Form)

	if len(states) != 1 {
		log.Fatal("asdf")
	}

	log.Println(states[0])
	s := &SearchState{}
	err := s.Decode(states[0])
	if err != nil {
		log.Fatal(err)
	}

	v := frame.Video{Filename: s.FileName}

	for i, x := range s.Segments {
		err := v.Split(x.Start, x.End, fmt.Sprintf("part%d.mp4", i))
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println(s)

	fmt.Fprintf(w, "Splitting...")
}

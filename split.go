package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func handleSplit(w http.ResponseWriter, r *http.Request) {
	log.Println("got split request!")

	r.ParseForm()
	states := r.Form["state"]
	log.Println(r.Form)

	_, finalize := r.Form["finalize"]

	if len(states) != 1 {
		log.Fatal("asdf")
	}

	log.Println(states[0])
	s := &SearchState{}
	err := s.Decode(states[0])
	if err != nil {
		log.Fatal(err)
	}

	v, err := VideoFrameCache.GetFrame(s.FileName)
	if err != nil {
		log.Fatal(err)
	}

	for i, x := range s.Segments {
		log.Println(i, x.Start, x.End)
		cachedLoc, err := v.Split(x.Start, x.End)
		if err != nil {
			log.Fatal(err)
		}

		if finalize {

			if _, err := os.Stat(finalizeFolderName(v.HashString)); os.IsNotExist(err) {
				err := os.Mkdir(v.HashString[:7], 0744)
				if err != nil {
					log.Fatal(err)
				}
			}

			err = copyFile(cachedLoc, finalizeName(i, v.HashString, x))
			if err != nil {
				log.Fatal(err)
			}
		}

	}

	log.Println(s)

	fmt.Fprintf(w, "Splitting...")
}

func finalizeName(i int, hash string, x Segment) string {
	label := x.Label
	if label == "" {
		label = "no label"
	}
	label = strings.Replace(label, " ", "-", -1)

	return fmt.Sprintf("%s/%02d_%s_%s.mp4", hash[:7], i, hash[:7], label)
}

func finalizeFolderName(hash string) string {
	return hash[:7]
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}

package main

import (
	"log"
	"net/http"
)

const (
	frameRoute = "/frame"
	splitRoute = "/split"
	port       = ":50181"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc(frameRoute, handleFrame)
	http.HandleFunc("/compare", handleCompare)
	http.HandleFunc("/done", handleDone)
	http.HandleFunc(splitRoute, handleSplit)

	log.Println("serving on port", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

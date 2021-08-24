package main

import (
	"log"
	"net/http"
)

const (
	compareRoute = "/compare"
	frameRoute   = "/frame"
	previewRoute = "/preview"
	splitRoute   = "/split"

	port = ":8080"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc(frameRoute, handleFrame)
	http.HandleFunc(previewRoute, handlePreview)
	http.HandleFunc(compareRoute, handleCompare)
	http.HandleFunc("/done", handleDone)
	http.HandleFunc(splitRoute, handleSplit)

	log.Println("serving on port", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

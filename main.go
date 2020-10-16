package main

import (
	"log"
	"net/http"
)

const (
	frameRoute = "/frame"
	splitRoute = "/split"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc(frameRoute, handleFrame)
	http.HandleFunc("/compare", handleCompare)
	http.HandleFunc("/done", handleDone)
	http.HandleFunc(splitRoute, handleSplit)

	port := ":8080"
	log.Println("serving on port", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

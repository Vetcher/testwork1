package main

import (
	"net/http"
	"log"
	"testwork1/src"
)

func main() {
	http.HandleFunc("/", ppp.MainHandler)
	log.Print("Listen on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
	close(ppp.EmailChannel)
}

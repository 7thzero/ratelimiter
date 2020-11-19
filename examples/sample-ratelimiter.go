package main

import (
	"net/http"
	"github.com/7thzero/ratelimiter/examples/samplerouter"
)

func main() {
	router := samplerouter.SampleRouter{}
	router.Init()

	http.HandleFunc("/", router.Route)
	http.ListenAndServe(":8421", nil)
}

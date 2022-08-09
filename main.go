package main

import (
	"log"
	"net/http"
	"time"

	docopt "github.com/docopt/docopt-go"
	"github.com/gorilla/mux"
)

// TODO: get these from config or environment variable
const (
	secret    = "verySecretSecret"
	rune      = "P9D3vkZZIawOf5YTRSt95Sdj2z9q8HiwuhAvNqaQKQY9MSZtZXRob2Q9aW52b2ljZQ=="
	lnHost    = "10.13.13.2:9735"
	lnNodeId  = "02b02f856f28cbe658133008b9dcb9ae2e6c18d27fbe5cd6644b6f13bcb42a269c"
	address   = "0.0.0.0:3333"
	staticDir = "/s"
	authPath  = "/s/books/"
)

const USAGE = `website
Usage:
  website run 
  website new-user <username>
`

var router = mux.NewRouter()

func main() {
	opts, err := docopt.ParseDoc(USAGE)
	if err != nil {
		return
	}

	switch {
	case opts["run"].(bool):
		runServer()
	case opts["new-user"].(bool):
		createUser(opts)
	}

	return
}

func runServer() {
	// all requests flow through authentication check and logger
	router.Use(authAndLogMiddleware)

	router.Path("/.well-known/lnurlp/{username}").Methods("GET").HandlerFunc(handleLNAddress)

	//	api := router.PathPrefix("/api").Subrouter()

	router.PathPrefix("/s/").Handler(http.StripPrefix("/s/", http.FileServer(http.Dir(staticDir))))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

	srv := &http.Server{
		Handler:      router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Listening on %v...\n", address)
	log.Fatal(srv.ListenAndServe())
}

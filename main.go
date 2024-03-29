package main

import (
	"log"
	"net/http"
	"time"

	docopt "github.com/docopt/docopt-go"
	"github.com/gorilla/mux"
)

// TODO: get these from config or environment variable
// test
const (
	secret    = "verySecretSecret"
	address   = "0.0.0.0:3333"
	domain    = "raph.ee"
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
	router.Use(logMiddleware)
	// router.Use(authMiddleware)

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

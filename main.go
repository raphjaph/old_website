package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	docopt "github.com/docopt/docopt-go"
	"github.com/gorilla/mux"
)

// TODO: get these from config or environment variable
const (
	secret    = "verySecretSecretJzT+pAqbdYMJAWZonC5Z"
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

func authAndLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// TODO: rate limit
		// only authenticate path to the actual book files
		if strings.HasPrefix(req.URL.Path, authPath) {
			username, password, ok := req.BasicAuth()
			if ok {
				expectedPassword := computePassword(username)

				// was reading about side channel attacks so I overengineered this password check
				if subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) == 1 {
					log.Printf("successfully authenticated %v", username)
					next.ServeHTTP(w, req)
					return
				}

				log.Printf("failed login attempt from %v for resource %v\n", req.RemoteAddr, req.RequestURI)
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return

			}
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		} else {
			// TODO: some logging?
			log.Println(req.URL)
		}

		// serve next if not trying to access /books/
		next.ServeHTTP(w, req)
		return
	})
}

// only creates a sort unique key (password) for the user; nothing stored in a database
func createUser(opts docopt.Opts) {
	username, ok := opts["<username>"].(string)
	if !ok {
		fmt.Println("error parsing commandline argument")
		return
	}

	fmt.Println(computePassword(username))
}

// computes a password for a username
// password = hash(username || secret); secret set in env or config
func computePassword(name string) string {
	name = strings.ToLower(name)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(name))
	return hex.EncodeToString(mac.Sum(nil))
}

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

	docopt "github.com/docopt/docopt-go"
)

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println(req.URL)
		next.ServeHTTP(w, req)
		return
	})
}

func authMiddleware(next http.Handler) http.Handler {
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

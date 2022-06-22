package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	docopt "github.com/docopt/docopt-go"
	"github.com/gorilla/mux"
	lnsocket "github.com/jb55/lnsocket/go"
	"github.com/tidwall/gjson"
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

	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/getinvoice", GetInvoice)

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

/*
    Old idea:
	https://raph.8el.eu/api/getinvoice?amount=999&description="invoice from website"
*/
func GetInvoice(writer http.ResponseWriter, req *http.Request) {
	//	values := req.URL.Query()
	//	amount := values.Get("amount")
	//	description := values.Get("description")

	invoice, err := lnSocketInvoice()
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
	}

	writer.Write([]byte(invoice))
	writer.WriteHeader(http.StatusOK)
}

func lnSocketInvoice() (string, error) {
	ln := lnsocket.LNSocket{}
	ln.GenKey()

	err := ln.ConnectAndInit(lnHost, lnNodeId)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer ln.Disconnect()

	label := "makeinvoice/" + strconv.FormatInt(time.Now().Unix(), 16)
	description := "from website tip jar"

	//params := fmt.Sprintf("[\"%dmsat\", \"%s\", \"%s\"]", mSatoshi, label, description)
	// any amount invoices
	params := fmt.Sprintf("[\"any\", \"%s\", \"%s\"]", label, description)

	body, err := ln.Rpc(rune, "invoice", params)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	resErr := gjson.Get(body, "error")
	if resErr.Type != gjson.Null {
		if resErr.Type == gjson.JSON {
			return "", errors.New(resErr.Get("message").String())
		} else if resErr.Type == gjson.String {
			return "", errors.New(resErr.String())
		}
		return "", fmt.Errorf("Unknown commando error: '%v'", resErr)
	}

	invoice := gjson.Get(body, "result.bolt11")
	if invoice.Type != gjson.String {
		return "", fmt.Errorf("No bolt11 result found in invoice response, got %v", body)
	}

	return invoice.String(), nil
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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	lnurl "github.com/fiatjaf/go-lnurl"
	"github.com/gorilla/mux"
	lnsocket "github.com/jb55/lnsocket/go"
	"github.com/tidwall/gjson"
)

const (
	rune     = "P9D3vkZZIawOf5YTRSt95Sdj2z9q8HiwuhAvNqaQKQY9MSZtZXRob2Q9aW52b2ljZQ=="
	lnHost   = "10.13.13.2:9735"
	lnNodeId = "02b02f856f28cbe658133008b9dcb9ae2e6c18d27fbe5cd6644b6f13bcb42a269c"
)

func handleLNAddress(writer http.ResponseWriter, req *http.Request) {
	username := mux.Vars(req)["username"]

	label := username + "@" + domain + "/" + strconv.FormatInt(time.Now().Unix(), 16)

	// minimum of info needed for ln address
	metadata := make([]interface{}, 0, 5)
	metadata = append(metadata, []string{"text/plain", "sending sats"})
	metadata = append(metadata, []string{"text/identifier", username + "@" + domain})

	enc, _ := json.Marshal(metadata)
	stringMetadata := string(enc)

	// 1. get the info
	if amount := req.URL.Query().Get("amount"); amount == "" {
		json.NewEncoder(writer).Encode(lnurl.LNURLPayResponse1{
			MinSendable:     1000,
			MaxSendable:     1000000000,
			Tag:             "payRequest",
			EncodedMetadata: stringMetadata,
			Callback:        fmt.Sprintf("https://%s/.well-known/lnurlp/%s", domain, username),
		})

		// 2. get the invoice
	} else {
		msat, err := strconv.Atoi(amount)
		if err != nil {
			json.NewEncoder(writer).Encode(lnurl.ErrorResponse("amount is not an integer"))
			return
		}

		bolt11, err := lnSocketInvoice(msat, label, stringMetadata, true)
		if err != nil {
			json.NewEncoder(writer).Encode(
				lnurl.ErrorResponse("failed to create invoice: " + err.Error()))
			return
		}

		json.NewEncoder(writer).Encode(lnurl.LNURLPayResponse2{
			PR:            bolt11,
			Routes:        make([][]lnurl.RouteInfo, 0),
			SuccessAction: lnurl.Action("magic internet money ftw ⚡️", ""),
		})
	}
}

func lnSocketInvoice(amount_msat int, label string, description string, useDescHash bool) (string, error) {
	ln := lnsocket.LNSocket{}
	ln.GenKey()

	err := ln.ConnectAndInit(lnHost, lnNodeId)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer ln.Disconnect()

	params := map[string]interface{}{
		"amount_msat": amount_msat,
		"label":       label,
		"description": description,
	}
	if useDescHash {
		params["deschashonly"] = true
	}

	json, _ := json.Marshal(params)
	stringParams := string(json)

	fmt.Println(stringParams)

	body, err := ln.Rpc(rune, "invoice", stringParams)
	if err != nil {
		fmt.Println(err)
		return "", err

	}

	fmt.Println(body)

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

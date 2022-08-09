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

func handleLNAddress(writer http.ResponseWriter, req *http.Request) {
	username := mux.Vars(req)["username"]

	// 1. get the info
	if amount := req.URL.Query().Get("amount"); amount == "" {
		json.NewEncoder(writer).Encode(lnurl.LNURLPayResponse1{
			LNURLResponse: lnurl.LNURLResponse{Status: "OK"},
			// Callback:       "https://raph.8el.eu/api/getinvoice",
			Callback:        fmt.Sprintf("https://raph.8el.eu/.well-known/lnurlp/%s", username),
			MinSendable:     1000,
			MaxSendable:     1000000000,
			EncodedMetadata: "",
			Tag:             "payRequest",
		})
		// 2. get the invoice
	} else {
		msat, err := strconv.Atoi(amount)
		if err != nil {
			json.NewEncoder(writer).Encode(lnurl.ErrorResponse("amount is not an integer"))
			return
		}

		label := "lnAddress/" + strconv.FormatInt(time.Now().Unix(), 16)
		description := "from LN Address"
		bolt11, err := lnSocketInvoice(msat, label, description)
		if err != nil {
			json.NewEncoder(writer).Encode(
				lnurl.ErrorResponse("failed to create invoice: " + err.Error()))
			return
		}

		json.NewEncoder(writer).Encode(lnurl.LNURLPayResponse2{
			LNURLResponse: lnurl.LNURLResponse{Status: "OK"},
			SuccessAction: lnurl.Action("Payment received!", ""),
			Routes:        make([][]lnurl.RouteInfo, 0),
			PR:            bolt11,
			Disposable:    lnurl.FALSE,
		})
	}
}

func lnSocketInvoice(msatoshi int, label string, description string) (string, error) {
	ln := lnsocket.LNSocket{}
	ln.GenKey()

	err := ln.ConnectAndInit(lnHost, lnNodeId)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer ln.Disconnect()

	// any amount invoices
	//params := fmt.Sprintf("[\"any\", \"%s\", \"%s\"]", label, description)
	params := fmt.Sprintf("[\"%d\", \"%s\", \"%s\"]", msatoshi, label, description)

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

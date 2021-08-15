package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cockroachdb/pebble"
	"github.com/fiatjaf/go-lnurl"
	"github.com/gorilla/mux"
)

func handleLNURL(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]

	log.Info().Str("username", username).Msg("got lnurl request")

	var params Params
	val, closer, err := db.Get([]byte(username))
	if err != nil {
		if err != pebble.ErrNotFound {
			log.Error().Err(err).Str("name", username).
				Msg("error getting data")
		}
		return
	}
	defer closer.Close()
	if err := json.Unmarshal(val, &params); err != nil {
		log.Error().Err(err).Str("name", username).Str("data", string(val)).
			Msg("got broken json from db")
		return
	}

	params.Name = username

	if amount := r.URL.Query().Get("amount"); amount == "" {
		// check if the receiver accepts comments
		var commentLength int64 = 0
		// TODO: support webhook comments

		json.NewEncoder(w).Encode(lnurl.LNURLPayResponse1{
			LNURLResponse:   lnurl.LNURLResponse{Status: "OK"},
			Callback:        fmt.Sprintf("https://%s/.well-known/lnurlp/%s", s.Domain, username),
			MinSendable:     1000,
			MaxSendable:     100000000,
			EncodedMetadata: makeMetadata(params),
			CommentAllowed:  commentLength,
			Tag:             "payRequest",
		})

	} else {
		msat, err := strconv.Atoi(amount)
		if err != nil {
			json.NewEncoder(w).Encode(lnurl.ErrorResponse("amount is not integer"))
			return
		}

		bolt11, err := makeInvoice(params, msat)
		if err != nil {
			json.NewEncoder(w).Encode(
				lnurl.ErrorResponse("failed to create invoice: " + err.Error()))
			return
		}

		json.NewEncoder(w).Encode(lnurl.LNURLPayResponse2{
			LNURLResponse: lnurl.LNURLResponse{Status: "OK"},
			PR:            bolt11,
			Routes:        make([][]lnurl.RouteInfo, 0),
			Disposable:    lnurl.FALSE,
			SuccessAction: lnurl.Action("Payment received!", ""),
		})

		// send webhook
		go func() {
			// TODO
		}()
	}
}

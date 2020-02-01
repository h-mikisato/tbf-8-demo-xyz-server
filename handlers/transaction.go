package handlers

import (
	"encoding/json"
	"net/http"

	jose "gopkg.in/square/go-jose.v2"
)

const (
	SignatureHeader = "JWS-Signature"
)

type request struct {
	Resources []json.RawMessage
	Keys      *keys
	Interact  *interact
	Display   []json.RawMessage
}

type keys struct {
	Proof string
	JWKs  *jwks
}

type interact struct {
	Redirect bool
	UserCode bool
	Callback *callback
}

type jwks struct {
	Keys []jose.JSONWebKey
}

type callback struct {
	uri   string
	nonce string
}

type response struct {
	InteractionURL string    `json:",omitempty"`
	ServerNonce    string    `json:",omitempty"`
	Wait           int       `json:",omitempty"`
	UserCode       *usercode `json:",omitempty"`
	Handle         *handle   `json:",omitempty"`
}

type handle struct {
	Value string
	Type  string
}

type usercode struct {
	URL  string
	Code string
}

type TransactionHandler struct {
	InteractionHost string
}

func (h *TransactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: verify RC's possession of key
	// sign := r.Header.Get(SignatureHeader)

	var res response
	if req.Interact.Redirect {
		interactionSeed := getHandle()
		res.InteractionURL = "https://" + h.InteractionHost + "/interact/" + interactionSeed
		serverNonce := getNonce()
		res.ServerNonce = serverNonce
	}
	if req.Interact.UserCode {
		res.UserCode = &usercode{
			URL:  "https://" + h.InteractionHost + "/interact/device",
			Code: getUserCode(),
		}
	}
	res.Handle = &handle{
		Value: getHandle(),
		Type:  "Bearer",
	}

	response, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

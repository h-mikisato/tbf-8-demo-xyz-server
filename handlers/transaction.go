package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	jose "gopkg.in/square/go-jose.v2"

	"cryptic-command/gatewatch/models"
)

const (
	SignatureHeader = "JWS-Signature"
)

type TransactionHandler struct {
	InteractionHost string
}

func (h *TransactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req models.Request
	if err := json.Unmarshal(payload, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sign := r.Header.Get(SignatureHeader)
	jws, err := jose.ParseDetached(sign, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := jws.Verify(req.Keys.JWKs.Keys[0]); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var res models.Response
	if req.Interact.Redirect {
		interactionSeed := getHandle()
		res.InteractionURL = "https://" + h.InteractionHost + "/interact/" + interactionSeed
		serverNonce := getNonce()
		res.ServerNonce = serverNonce
	}
	if req.Interact.UserCode {
		res.UserCode = &models.Usercode{
			URL:  "https://" + h.InteractionHost + "/interact/device",
			Code: getUserCode(),
		}
	}
	res.Handle = &models.Handle{
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

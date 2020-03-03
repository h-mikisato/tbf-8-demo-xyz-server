package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	jose "gopkg.in/square/go-jose.v2"

	"cryptic-command/gatewatch/models"
	"cryptic-command/gatewatch/repositories"
)

const (
	SignatureHeader = "JWS-Signature"
	BearerTokenType = "Bearer"
	WaitInterval    = 30
)

type TransactionHandler struct {
	InteractionHost string
	Repository      *repositories.Transaction
}

func (h *TransactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// read error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req models.Request
	if err := json.Unmarshal(payload, &req); err != nil {
		// parse error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check possession of key
	// for easy, only detached jws
	sign := r.Header.Get(SignatureHeader)
	jws, err := jose.ParseDetached(sign, payload)
	if err != nil {
		// jws parse error
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := jws.Verify(req.Keys.JWKs.Keys[0]); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorMessage("user_denied"))
		return
	}

	if req.Handle == "" {
		h.firstTransaction(w, &req)
	} else {
		h.handleState(w, &req)
	}
}

func (h *TransactionHandler) firstTransaction(w http.ResponseWriter, req *models.Request) {
	var (
		res models.Response
		t   = models.NewTransaction()
	)

	if req.Interact.Redirect {
		// Redirect with Callback / Redirect with Polling
		t.Handle = getRandomB32(15)
		t.InteractionKey = getRandomB32(15)
		t.State = models.WaitingForAuthz
		t.InteractionType = models.RedirectInteraction
		res.InteractionURL = "https://" + h.InteractionHost + "/interact/" + t.InteractionKey
		if req.Interact.Callback != nil {
			// Redirect with Callback only
			t.ServerNonce = getRandomB32(20)
			t.ResponseURL = req.Interact.Callback.URI
			t.ClientNonce = req.Interact.Callback.Nonce
			res.ServerNonce = t.ServerNonce
		}
	}
	if req.Interact.UserCode {
		// UserCode with Polling
		// not implemented
	}
	res.Handle = &models.Token{
		Value: t.Handle,
		Type:  "Bearer",
	}

	response, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.Repository.Update(t, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *TransactionHandler) handleState(w http.ResponseWriter, req *models.Request) {
	var (
		t   *models.Transaction
		err error
	)
	t, err = h.Repository.Get(req.Handle)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorMessage("unknown_transaction"))
		return
	}

	if !compareKey(t.Key, req.Keys.JWKs.Keys[0]) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorMessage("user_denied"))
		return
	}

	if t.IsExpired(time.Now().UTC()) {
		h.Repository.Drop(t)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorMessage("expired"))
		return
	}

	var (
		res       models.Response
		response  []byte
		oldHandle = t.Handle
	)
	switch t.State {
	case models.WaitingForAuthz:
		t.Handle = getRandomB32(15)
		res.Wait = WaitInterval
		res.Handle = &models.Token{
			Value: t.Handle,
			Type:  BearerTokenType,
		}

	case models.WaitingForIssuing:
		if req.InteractRef != t.InteractionRef {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errorMessage("user_denied"))
			return
		}
		t.Handle = getRandomB32(15)
		t.State = models.Issued
		res.Handle = &models.Token{
			Value: t.Handle,
			Type:  BearerTokenType,
		}
		res.AccessToken = &models.Token{
			Value: getRandomB32(25),
			Type:  BearerTokenType,
		}

	case models.Issued:
		t.Handle = getRandomB32(15)
		res.Handle = &models.Token{
			Value: t.Handle,
			Type:  BearerTokenType,
		}
		res.AccessToken = &models.Token{
			Value: getRandomB32(25),
			Type:  BearerTokenType,
		}
	}

	response, err = json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Repository.Update(t, oldHandle)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

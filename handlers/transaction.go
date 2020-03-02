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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req models.Request
	if err := json.Unmarshal(payload, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check possession of key
	// for easy, only detached jws
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
		t.Handle = getHandle()
		t.InteractionKey = getHandle()
		t.State = models.WaitingForAuthz
		t.InteractionType = models.RedirectInteraction
		res.InteractionURL = "https://" + h.InteractionHost + "/interact/" + t.InteractionKey
		if req.Interact.Callback != nil {
			// Redirect with Callback only
			t.ServerNonce = getNonce()
			t.ResponseURL = req.Interact.Callback.URI
			t.ClientNonce = req.Interact.Callback.Nonce
			res.ServerNonce = t.ServerNonce
		}
	}
	if req.Interact.UserCode {
		t.Handle = getHandle()
		t.InteractionKey = getUserCode()
		t.InteractionType = models.UserCodeInteraction
		res.UserCode = &models.Usercode{
			URL:  "https://" + h.InteractionHost + "/interact/device",
			Code: t.InteractionKey,
		}
	}
	res.Handle = &models.Token{
		Value: t.Handle,
		Type:  "Bearer",
	}

	response, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !compareKey(t.Key, req.Keys.JWKs.Keys[0]) {
		http.Error(w, "transaction key is not match", http.StatusBadRequest)
		return
	}

	if t.IsExpired(time.Now().UTC()) {
		h.Repository.Drop(t)
		http.Error(w, "transaction is expired", http.StatusBadRequest)
		return
	}

	var (
		res       models.Response
		response  []byte
		oldHandle = t.Handle
	)
	switch t.State {
	case models.WaitingForAuthz:
		t.Handle = getHandle()
		res.Wait = WaitInterval
		res.Handle = &models.Token{
			Value: t.Handle,
			Type:  BearerTokenType,
		}

	case models.WaitingForIssuing:
		if req.InteractRef != t.InteractionRef {
			http.Error(w, "not match interact ref", http.StatusBadRequest)
			return
		}
		t.Handle = getHandle()
		t.State = models.Issued
		res.Handle = &models.Token{
			Value: t.Handle,
			Type:  BearerTokenType,
		}
		res.AccessToken = &models.Token{
			Value: getToken(),
			Type:  BearerTokenType,
		}

	case models.Issued:
		t.Handle = getHandle()
		res.Handle = &models.Token{
			Value: t.Handle,
			Type:  BearerTokenType,
		}
		res.AccessToken = &models.Token{
			Value: getToken(),
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

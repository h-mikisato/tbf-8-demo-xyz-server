package handlers

import (
	"crypto/sha512"
	"hash"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/sha3"

	"cryptic-command/gatewatch/models"
	"cryptic-command/gatewatch/repositories"
)

const (
	UserCodeInteractionPath = "device"
)

type InteractionHandler struct {
	Repository *repositories.Transaction
}

func (h *InteractionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handle := mux.Vars(r)["handle"]
	if handle == UserCodeInteractionPath {
		h.deviceHandler(w, r)
	} else {
		h.redirectHandler(w, r, handle)
	}
}

func (h *InteractionHandler) deviceHandler(w http.ResponseWriter, r *http.Request) {
	// UserCode with Polling
	// stub
}

func (h *InteractionHandler) redirectHandler(w http.ResponseWriter, r *http.Request, handle string) {
	t, err := h.Repository.GetFromInteraction(handle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if t.InteractionType != models.RedirectInteraction {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if t.IsExpired(time.Now().UTC()) {
		h.Repository.Drop(t)
		http.Error(w, "transaction is expired", http.StatusBadRequest)
		return
	}

	if t.State != models.WaitingForAuthz {
		http.Error(w, "transaction is not waiting authorization", http.StatusBadRequest)
		return
	}

	/////////////////////////////////////////////////////
	// 本来はここで認証ページを表示し、認証を受けつける。
	// 以下、認証できた場合の処理
	/////////////////////////////////////////////////////

	if t.ResponseURL == "" {
		// Redirect with Polling

		err := h.Repository.UpdateByInteraction(handle, models.WaitingForIssuing, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// Redirect with Callback

	responseURL, err := url.Parse(t.ResponseURL)
	if err != nil {
		h.Repository.Drop(t)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var hasher hash.Hash
	if t.HashAlgo == "sha2" {
		hasher = sha512.New()
	} else {
		hasher = sha3.New512()
	}

	interactionRef := getHandle()

	interactionHash := makeInteractionHash(t.ServerNonce, t.ClientNonce, interactionRef, hasher)

	err = h.Repository.UpdateByInteraction(handle, models.WaitingForIssuing, interactionRef)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := make(url.Values, 2)
	query.Add("hash", interactionHash)
	query.Add("interact", interactionRef)

	responseURL.RawQuery = query.Encode()

	w.Header().Add("Location", responseURL.String())
	w.WriteHeader(http.StatusFound)
}

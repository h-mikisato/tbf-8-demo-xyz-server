package handlers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/sha3"
)

const (
	UserCodeInteractionPath = "device"

	dummyServerNonce = "MBDOFXG4Y5CVJCX821LH"
	dummyClientNonce = "LKLTI25DK82FX4T4QFZC"
	dummyResponseURL = "https://client.example.net/return/123455"
)

type InteractionHandler struct {
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
	// mock
}

func (h *InteractionHandler) redirectHandler(w http.ResponseWriter, r *http.Request, handle string) {
	responseURL, _ := url.Parse(dummyResponseURL)
	query := make(url.Values, 2)

	interactionHandle := getHandle()
	hasher := sha3.New512()
	interactionHash := makeInteractionHash(dummyServerNonce, dummyClientNonce, interactionHandle, hasher)

	query.Add("hash", interactionHash)
	query.Add("interaction_handle", interactionHandle)

	responseURL.RawQuery = query.Encode()

	w.Header().Add("Location", responseURL.String())
	w.WriteHeader(http.StatusFound)
}

package models

import (
	"encoding/json"

	"gopkg.in/square/go-jose.v2"
)

type Request struct {
	Resources []json.RawMessage
	Keys      *Keys
	Interact  *Interact
	Display   json.RawMessage
}

type Keys struct {
	Proof string
	JWKs  *JWKs
}

type Interact struct {
	Redirect bool
	UserCode bool
	Callback *Callback
}

type JWKs struct {
	Keys []jose.JSONWebKey
}

type Callback struct {
	URI   string
	Nonce string
}

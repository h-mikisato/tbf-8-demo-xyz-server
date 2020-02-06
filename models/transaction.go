package models

import (
	"time"

	"gopkg.in/square/go-jose.v2"
)

type TransactionStatus int

type Transaction struct {
	Status      TransactionStatus
	ServerNonce string
	ClientNonce string
	HashAlgo    string
	ResponseURL string
	UserCode    string
	Key         jose.JSONWebKey
	LastUpdated time.Time
}

const (
	Initialized TransactionStatus = iota
	WaitingForAuthz
	WaitingForIssuing
	Issued
)

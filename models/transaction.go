package models

import (
	"time"

	"gopkg.in/square/go-jose.v2"
)

const (
	transactionTimeout = time.Duration(time.Hour * 24 * 30)
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

func (t *Transaction) IsExpired(now time.Time) bool {
	return now.Sub(t.LastUpdated) > transactionTimeout
}

func NewTransaction() *Transaction {
	return &Transaction{
		Status:      Initialized,
		LastUpdated: time.Now().UTC(),
	}
}

const (
	Initialized TransactionStatus = iota
	WaitingForAuthz
	WaitingForIssuing
	Issued
)

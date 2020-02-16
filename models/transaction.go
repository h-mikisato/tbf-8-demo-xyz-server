package models

import (
	"time"

	"github.com/jinzhu/copier"
	"gopkg.in/square/go-jose.v2"
)

const (
	transactionTimeout = time.Duration(time.Hour * 24 * 30)
)

type (
	TransactionState int
	InteractionType  int
)

const (
	Initialized TransactionState = iota
	WaitingForAuthz
	WaitingForIssuing
	Issued
)

const (
	NoInteraction InteractionType = iota
	RedirectInteraction
	UserCodeInteraction
)

type Transaction struct {
	Handle          string
	State           TransactionState
	InteractionType InteractionType
	ServerNonce     string
	ClientNonce     string
	HashAlgo        string
	ResponseURL     string
	InteractionKey  string // user_code or interaction URL unique key
	InteractionRef  string
	Key             jose.JSONWebKey
	LastUpdated     time.Time
}

func (t *Transaction) IsExpired(now time.Time) bool {
	return now.Sub(t.LastUpdated) > transactionTimeout
}

func (t *Transaction) Clone() *Transaction {
	that := &Transaction{}
	copier.Copy(that, t)
	return that
}

func NewTransaction() *Transaction {
	return &Transaction{
		LastUpdated: time.Now().UTC(),
	}
}

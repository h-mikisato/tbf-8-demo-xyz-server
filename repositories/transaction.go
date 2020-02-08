package repositories

import (
	"errors"
	"sync"

	"cryptic-command/gatewatch/models"
)

var (
	ErrTransactionNotExists = errors.New("transaction not exists")
)

type Transaction struct {
	sync.RWMutex
	pool        map[string]*models.Transaction
	interactKey map[string]string
}

func NewTransaction() *Transaction {
	return &Transaction{
		pool:        make(map[string]*models.Transaction),
		interactKey: make(map[string]string),
	}
}

func (repo *Transaction) Get(handle string) (*models.Transaction, error) {
	repo.RLock()
	defer repo.RUnlock()
	t, ok := repo.pool[handle]
	if !ok {
		return nil, ErrTransactionNotExists
	}
	return t.Clone(), nil
}

func (repo *Transaction) GetFromInteraction(interaction string) (*models.Transaction, error) {
	repo.RLock()
	defer repo.RUnlock()
	handle, ok := repo.interactKey[interaction]
	if !ok {
		return nil, ErrTransactionNotExists
	}
	t, ok := repo.pool[handle]
	if !ok {
		return nil, ErrTransactionNotExists
	}
	return t.Clone(), nil
}

func (repo *Transaction) Store(t *models.Transaction, oldHandle string) {
	repo.Lock()
	defer repo.Unlock()
	repo.pool[t.Handle] = t
	if t.InteractionKey != "" {
		repo.interactKey[t.InteractionKey] = t.Handle
	}
}

func (repo *Transaction) Drop(t *models.Transaction) {
	repo.Lock()
	defer repo.Unlock()
	delete(repo.pool, t.Handle)
	if t.InteractionKey != "" {
		delete(repo.interactKey, t.InteractionKey)
	}
}

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
	pool map[string]*models.Transaction
}

func (repo *Transaction) Get(handle string) (*models.Transaction, error) {
	repo.RLock()
	defer repo.RUnlock()
	t, ok := repo.pool[handle]
	if ok {
		return t, nil
	}
	return nil, ErrTransactionNotExists
}

func (repo *Transaction) Store(oldHandle, handle string, t *models.Transaction) {
	repo.Lock()
	defer repo.Unlock()
	if oldHandle != "" {
		delete(repo.pool, oldHandle)
	}
	repo.pool[handle] = t
}

func (repo *Transaction) Drop(handle string) {
	repo.Lock()
	defer repo.Unlock()
	delete(repo.pool, handle)
}

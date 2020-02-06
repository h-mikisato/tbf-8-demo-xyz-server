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
	sync.Mutex
	pool     map[string]*models.Transaction
	userCode map[string]string
}

func (repo *Transaction) Get(handle string) (*models.Transaction, error) {
	repo.Lock()
	defer repo.Unlock()
	t, ok := repo.pool[handle]
	if !ok {
		return nil, ErrTransactionNotExists
	}
	delete(repo.pool, handle)
	return t, nil
}

func (repo *Transaction) GetFromUserCode(userCode string) (*models.Transaction, error) {
	repo.Lock()
	defer repo.Unlock()
	handle, ok := repo.userCode[userCode]
	if !ok {
		return nil, ErrTransactionNotExists
	}
	delete(repo.userCode, userCode)
	t, ok := repo.pool[handle]
	if !ok {
		return nil, ErrTransactionNotExists
	}
	delete(repo.pool, handle)
	return t, nil
}

func (repo *Transaction) Store(handle string, t *models.Transaction) {
	repo.Lock()
	defer repo.Unlock()
	repo.pool[handle] = t
	if t.UserCode != "" {
		repo.userCode[t.UserCode] = handle
	}
}

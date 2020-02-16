package repositories

import (
	"errors"
	"sync"

	"cryptic-command/gatewatch/models"
)

var (
	ErrInvalidTransaction = errors.New("invalid transaction")
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
		return nil, ErrInvalidTransaction
	}
	return t.Clone(), nil
}

func (repo *Transaction) GetFromInteraction(interaction string) (*models.Transaction, error) {
	repo.RLock()
	defer repo.RUnlock()
	handle, ok := repo.interactKey[interaction]
	if !ok {
		return nil, ErrInvalidTransaction
	}
	t, ok := repo.pool[handle]
	if !ok {
		return nil, ErrInvalidTransaction
	}
	return t.Clone(), nil
}

func (repo *Transaction) Update(t *models.Transaction, oldHandle string) error {
	repo.Lock()
	defer repo.Unlock()
	if oldHandle != "" {
		_, ok := repo.pool[oldHandle]
		if !ok {
			return ErrInvalidTransaction
		}
		delete(repo.pool, oldHandle)
	}
	repo.pool[t.Handle] = t
	if t.InteractionKey != "" {
		repo.interactKey[t.InteractionKey] = t.Handle
	}
	return nil
}

func (repo *Transaction) UpdateByInteraction(interaction string, state models.TransactionState, ref string) error {
	repo.Lock()
	defer repo.Unlock()
	handle, ok := repo.interactKey[interaction]
	if !ok {
		return ErrInvalidTransaction
	}
	t, ok := repo.pool[handle]
	if !ok {
		return ErrInvalidTransaction
	}
	t.State = state
	if ref != "" {
		t.InteractionRef = ref
	}

	delete(repo.interactKey, handle)
	return nil
}

func (repo *Transaction) Drop(t *models.Transaction) {
	repo.Lock()
	defer repo.Unlock()
	delete(repo.pool, t.Handle)
	if t.InteractionKey != "" {
		delete(repo.interactKey, t.InteractionKey)
	}
}

package inmemory

import (
	"errors"
	"sync"

	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

var errTransactionClosed = errors.New("transaction is closed")

var _ apputxo.Repository = (*Block)(nil)

type Block struct {
	mu sync.RWMutex
	db map[utxo.ID]utxo.UTXO
}

func NewBlock() *Block {
	return &Block{db: make(map[utxo.ID]utxo.UTXO)}
}

func (r *Block) FindByID(id utxo.ID) (utxo.UTXO, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.db[id]
	return u, exists, nil
}

func (r *Block) FindByOwner(owner types.PublicKey) ([]utxo.UTXO, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	utxos := make([]utxo.UTXO, 0, len(r.db))

	for _, u := range r.db {
		if u.Output.Owner.Equal(owner.PublicKey) {
			utxos = append(utxos, u)
		}
	}

	return utxos, nil
}

func (r *Block) Save(u utxo.UTXO) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.db[u.ID] = u
	return nil
}

func (r *Block) Delete(id utxo.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.db, id)
	return nil
}

func (r *Block) Begin() (apputxo.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	newMap := make(map[utxo.ID]utxo.UTXO, len(r.db))

	for k, v := range r.db {
		newMap[k] = v
	}

	return &transaction{
		repository: r,
		db:         newMap,
		closed:     false,
	}, nil
}

type transaction struct {
	repository *Block
	db         map[utxo.ID]utxo.UTXO
	closed     bool
}

func (t *transaction) FindByID(id utxo.ID) (utxo.UTXO, bool, error) {
	t.repository.mu.RLock()
	defer t.repository.mu.RUnlock()

	if t.closed {
		return utxo.UTXO{}, false, errTransactionClosed
	}

	u, exists := t.db[id]
	return u, exists, nil
}

func (t *transaction) FindByOwner(owner types.PublicKey) ([]utxo.UTXO, error) {
	t.repository.mu.RLock()
	defer t.repository.mu.RUnlock()

	if t.closed {
		return []utxo.UTXO{}, errTransactionClosed
	}

	utxos := make([]utxo.UTXO, 0, len(t.db))

	for _, u := range t.db {
		if u.Output.Owner.Equal(owner) {
			utxos = append(utxos, u)
		}
	}

	return utxos, nil
}

func (t *transaction) Save(u utxo.UTXO) error {
	t.repository.mu.Lock()
	defer t.repository.mu.Unlock()

	if t.closed {
		return errTransactionClosed
	}

	t.db[u.ID] = u
	return nil
}

func (t *transaction) Delete(id utxo.ID) error {
	t.repository.mu.Lock()
	defer t.repository.mu.Unlock()

	if t.closed {
		return errTransactionClosed
	}

	delete(t.db, id)
	return nil
}

func (t *transaction) Commit() error {
	t.repository.mu.Lock()
	defer t.repository.mu.Unlock()

	if t.closed {
		return errTransactionClosed
	}

	t.repository.db = t.db
	t.closed = true

	return nil
}

func (t *transaction) Rollback() error {
	if !t.closed {
		t.closed = true
	}

	return nil
}

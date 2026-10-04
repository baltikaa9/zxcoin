package inmemory

import (
	"crypto/ecdsa"
	"errors"

	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

var errTransactionClosed = errors.New("transaction is closed")

type Repository struct {
	db map[utxo.UTXOID]utxo.UTXO
}

func NewRepository() *Repository {
	return &Repository{db: make(map[utxo.UTXOID]utxo.UTXO)}
}

func (r *Repository) FindByID(id utxo.UTXOID) (utxo.UTXO, bool, error) {
	u, exists := r.db[id]
	return u, exists, nil
}

func (r *Repository) FindByOwner(owner *ecdsa.PublicKey) ([]utxo.UTXO, error) {
	utxos := make([]utxo.UTXO, 0, len(r.db))

	for _, u := range r.db {
		if u.Output.Owner.Equal(owner) {
			utxos = append(utxos, u)
		}
	}

	return utxos, nil
}

func (r *Repository) Save(u utxo.UTXO) error {
	r.db[u.ID] = u
	return nil
}

func (r *Repository) Delete(id utxo.UTXOID) error {
	delete(r.db, id)
	return nil
}

func (r *Repository) Begin() (apputxo.Transaction, error) {
	newMap := make(map[utxo.UTXOID]utxo.UTXO, len(r.db))

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
	repository *Repository
	db         map[utxo.UTXOID]utxo.UTXO
	closed     bool
}

func (t *transaction) FindByID(id utxo.UTXOID) (utxo.UTXO, bool, error) {
	if t.closed {
		return utxo.UTXO{}, false, errTransactionClosed
	}

	u, exists := t.db[id]
	return u, exists, nil
}

func (t *transaction) FindByOwner(owner *ecdsa.PublicKey) ([]utxo.UTXO, error) {
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
	if t.closed {
		return errTransactionClosed
	}

	t.db[u.ID] = u
	return nil
}

func (t *transaction) Delete(id utxo.UTXOID) error {
	if t.closed {
		return errTransactionClosed
	}

	delete(t.db, id)
	return nil
}

func (t *transaction) Commit() error {
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

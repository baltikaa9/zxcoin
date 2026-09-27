package inmemory

import (
	"crypto/ecdsa"
	"zxcoin/utxo"
)

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
		if u.Owner.Equal(owner) {
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

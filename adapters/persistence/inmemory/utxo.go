package inmemory

import (
	"sync"

	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

var _ apputxo.Repository = (*UTXO)(nil)

type UTXO struct {
	mu sync.RWMutex
	db map[utxo.ID]utxo.UTXO
}

func NewUTXO() *UTXO {
	return &UTXO{db: make(map[utxo.ID]utxo.UTXO)}
}

func (r *UTXO) FindByID(id utxo.ID) (utxo.UTXO, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.db[id]
	return u, exists, nil
}

func (r *UTXO) FindByOwner(owner types.PublicKey) ([]utxo.UTXO, error) {
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

func (r *UTXO) Save(u utxo.UTXO) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.db[u.ID] = u
}

func (r *UTXO) Delete(id utxo.ID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.db, id)
}

package inmemory

import (
	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

var _ apputxo.Repository = (*UTXO)(nil)

type UTXO struct {
	db map[utxo.ID]utxo.UTXO
}

func NewUTXO() *UTXO {
	return &UTXO{db: make(map[utxo.ID]utxo.UTXO)}
}

func (r *UTXO) FindByID(id utxo.ID) (utxo.UTXO, bool, error) {
	u, exists := r.db[id]
	return u, exists, nil
}

func (r *UTXO) FindByOwner(owner types.PublicKey) ([]utxo.UTXO, error) {
	utxos := make([]utxo.UTXO, 0, len(r.db))

	for _, u := range r.db {
		if u.Output.Owner.Equal(owner.PublicKey) {
			utxos = append(utxos, u)
		}
	}

	return utxos, nil
}

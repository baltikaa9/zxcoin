package utxo

import (
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type Repository interface {
	FindByID(id utxo.ID) (utxo.UTXO, bool, error)
	FindByOwner(owner types.PublicKey) ([]utxo.UTXO, error)
}

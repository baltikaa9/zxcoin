package utxo

import (
	"crypto/ecdsa"

	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type Repository interface {
	FindByID(id utxo.UTXOID) (utxo.UTXO, bool, error)
	FindByOwner(owner *ecdsa.PublicKey) ([]utxo.UTXO, error)
	Save(utxo utxo.UTXO) error
	Delete(id utxo.UTXOID) error

	Begin() (Transaction, error)
}

type Transaction interface {
	FindByID(id utxo.UTXOID) (utxo.UTXO, bool, error)
	FindByOwner(owner *ecdsa.PublicKey) ([]utxo.UTXO, error)
	Save(utxo utxo.UTXO) error
	Delete(id utxo.UTXOID) error

	Commit() error
	Rollback() error
}

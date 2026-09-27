package utxo

import "crypto/ecdsa"

type Repository interface {
	FindByID(id UTXOID) (UTXO, bool, error)
	FindByOwner(owner *ecdsa.PublicKey) ([]UTXO, error)
	Save(utxo UTXO) error
	Delete(id UTXOID) error
}

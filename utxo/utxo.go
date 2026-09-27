// Package utxo хранит набор непотраченных выходов транзакций (UTXO) и управляет их резервированием при создании новых транзакций.
package utxo

import (
	"crypto/ecdsa"
)

type UTXO struct {
	ID     UTXOID
	Amount uint64
	Owner  *ecdsa.PublicKey
}

type UTXOID struct {
	TxID     [32]byte
	OutIndex uint64
}

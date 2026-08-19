package utxo

import (
	"zxcoin/coin"
)

type UTXOKey struct {
	TxID     [32]byte
	OutIndex int
}

type UTXOEntry struct {
	Output   coin.TxOutput
	reserved bool
}

type UTXODB map[UTXOKey]UTXOEntry

func (e UTXOEntry) Reserved() bool {
	return e.reserved
}

func (db UTXODB) Reserve(key UTXOKey) error {
	entry, exists := db[key]

	if !exists {
		return &UTXOKeyNotFoundError{Key: key}
	}

	if entry.Reserved() {
		return &UTXOAlreadyReservedError{Key: key}
	}

	entry.reserved = true
	db[key] = entry

	return nil
}

func (db UTXODB) Release(key UTXOKey) {
	entry, exists := db[key]

	if !exists {
		return
	}

	entry.reserved = false
	db[key] = entry
}

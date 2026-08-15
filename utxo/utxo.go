package utxo

import "zxcoin/coin"

type UTXOKey struct {
	TxID     [32]byte
	OutIndex int
}

type UTXOEntry struct {
	Output   coin.TxOutput
	Reserved bool
}

type UTXODB map[UTXOKey]UTXOEntry

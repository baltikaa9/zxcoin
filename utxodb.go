package main

type UTXOKey struct {
	TxID     [32]byte
	OutIndex int
}

type UTXODB map[UTXOKey]TxOutput

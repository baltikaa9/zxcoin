package main

type UTXOKey struct {
	TxID     string
	OutIndex int
}

type UTXODB map[UTXOKey]TxOutput

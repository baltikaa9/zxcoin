package main

type UTXOKey struct {
	TxID     [32]byte
	OutIndex int
}

type UTXOEntry struct {
	Output   TxOutput
	Reserved bool
}

type UTXODB map[UTXOKey]UTXOEntry

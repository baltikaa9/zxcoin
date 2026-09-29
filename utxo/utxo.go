// Package utxo хранит набор непотраченных выходов транзакций (UTXO) и управляет их резервированием при создании новых транзакций.
package utxo

import (
	"zxcoin/coin"
)

type UTXOID struct {
	TxID     [32]byte
	OutIndex uint64
}

type UTXO struct {
	ID     UTXOID
	Output coin.TxOutput
}

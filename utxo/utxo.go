// Package utxo хранит набор непотраченных выходов транзакций (UTXO) и управляет их резервированием при создании новых транзакций.
package utxo

import (
	"github.com/baltikaa9/zxcoin/coin"
	"github.com/baltikaa9/zxcoin/types"
)

type UTXOID struct {
	TxID     types.Hash
	OutIndex uint64
}

type UTXO struct {
	ID     UTXOID
	Output coin.TxOutput
}

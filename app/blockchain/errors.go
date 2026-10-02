package blockchain

import (
	"fmt"

	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type DoubleSpendError struct {
	TxID     types.Hash
	OutIndex uint64
}

func (e DoubleSpendError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) уже используется в данном блоке", e.TxID, e.OutIndex)
}

type MoreOneCoinbaseError struct{}

func (e MoreOneCoinbaseError) Error() string {
	return "больше одной coinbase-транзакции в блоке"
}

package mempool

import (
	"fmt"

	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type DoubleSpendError struct {
	TxID     types.Hash
	OutIndex uint64
}

func (e DoubleSpendError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) уже добавлено в mempool", e.TxID, e.OutIndex)
}

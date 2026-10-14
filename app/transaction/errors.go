package transaction

import (
	"fmt"

	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type UTXONotFoundError struct {
	TxID     types.Hash
	OutIndex uint64
}

func (e UTXONotFoundError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) не найден", e.TxID, e.OutIndex)
}

type DuplicateInputError struct {
	TxID     types.Hash
	OutIndex uint64
}

func (e DuplicateInputError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) уже используется в данной транзакции", e.TxID, e.OutIndex)
}

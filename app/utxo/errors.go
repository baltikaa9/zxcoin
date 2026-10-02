package utxo

import (
	"fmt"

	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type UTXONotFoundError struct {
	ID utxo.UTXOID
}

func (e UTXONotFoundError) Error() string {
	return fmt.Sprintf("UTXO %v:%v не найден", e.ID.TxID, e.ID.OutIndex)
}

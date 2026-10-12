package utxo

import (
	"fmt"

	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type NotFoundError struct {
	ID utxo.ID
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("UTXO %v:%v не найден", e.ID.TxID, e.ID.OutIndex)
}

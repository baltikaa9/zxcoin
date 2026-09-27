package utxo

import "fmt"

type UTXONotFoundError struct {
	ID UTXOID
}

func (e UTXONotFoundError) Error() string {
	return fmt.Sprintf("UTXO %v:%v не найден", e.ID.TxID, e.ID.OutIndex)
}

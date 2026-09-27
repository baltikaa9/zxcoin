package utxo

import "fmt"

type UTXONotFoundError struct {
	ID UTXOID
	// TxID     [32]byte
	// OutIndex uint64
}

func (e UTXONotFoundError) Error() string {
	return fmt.Sprintf("UTXO %v:%v не найден", e.ID.TxID, e.ID.OutIndex)
}

// type UTXOAlreadyReservedError struct {
	// Key UTXOKey
// }

// func (e *UTXOAlreadyReservedError) Error() string {
	// return fmt.Sprintf("UTXO %v уже зарезервирована", e.Key)
// }

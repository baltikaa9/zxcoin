package utxo

import "fmt"

type UTXONotFoundError struct {
	ID UTXOID
	// TxID     [32]byte
	// OutIndex uint64
}

func (e *UTXOKeyNotFoundError) Error() string {
	return fmt.Sprintf("UTXO %v не найден", e.Key)
}

// type UTXOAlreadyReservedError struct {
	// Key UTXOKey
// }

func (e *UTXOAlreadyReservedError) Error() string {
	return fmt.Sprintf("UTXO %v уже зарезервирована", e.Key)
}

package transaction

import "fmt"

type UTXONotFoundError struct {
	TxID     [32]byte
	OutIndex int
}

func (e *UTXONotFoundError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) не найден", e.TxID, e.OutIndex)
}

type InvalidSignatureError struct {
	TxID     [32]byte
	OutIndex int
}

func (e *InvalidSignatureError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) не верная подпись", e.TxID, e.OutIndex)
}

type NotEnoughMoneyError struct {
	Input  int
	Output int
}

func (e *NotEnoughMoneyError) Error() string {
	return fmt.Sprintf("недостаточно средств: входы %d, выходы %d", e.Input, e.Output)
}

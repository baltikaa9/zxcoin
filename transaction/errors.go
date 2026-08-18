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

type InsufficientFundsError struct {
	Input  int
	Output int
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("недостаточно средств: входы %d, выходы %d", e.Input, e.Output)
}

type NonPositiveOutputError struct {
	TxID     [32]byte
	OutIndex int
	Value    int
}

func (e *NonPositiveOutputError) Error() string {
	return fmt.Sprintf("неположительное значение выхода %v транзакции %v: %v", e.OutIndex, e.TxID, e.Value)
}

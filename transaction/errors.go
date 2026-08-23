package transaction

import "fmt"

type UTXONotFoundError struct {
	TxID     [32]byte
	OutIndex uint64
}

func (e *UTXONotFoundError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) не найден", e.TxID, e.OutIndex)
}

type InvalidSignatureError struct {
	TxID     [32]byte
	OutIndex uint64
}

func (e *InvalidSignatureError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) не верная подпись", e.TxID, e.OutIndex)
}

type InsufficientFundsError struct {
	Input  uint64
	Output uint64
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("недостаточно средств: входы %d, выходы %d", e.Input, e.Output)
}

type ZeroOutputError struct {
	TxID     [32]byte
	OutIndex uint64
}

func (e *ZeroOutputError) Error() string {
	return fmt.Sprintf("нулевое значение выхода %v транзакции %v", e.OutIndex, e.TxID)
}

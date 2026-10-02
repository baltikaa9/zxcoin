package transaction

import (
	"fmt"

	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type InvalidSignatureError struct {
	TxID     types.Hash
	OutIndex uint64
}

func (e InvalidSignatureError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) не верная подпись", e.TxID, e.OutIndex)
}

type InsufficientFundsError struct {
	Input  uint64
	Output uint64
}

func (e InsufficientFundsError) Error() string {
	return fmt.Sprintf("недостаточно средств: входы %d, выходы %d", e.Input, e.Output)
}

type ZeroOutputError struct {
	TxID     types.Hash
	OutIndex uint64
}

func (e ZeroOutputError) Error() string {
	return fmt.Sprintf("нулевое значение выхода %v транзакции %v", e.OutIndex, e.TxID)
}

type EmptyOutputsError struct {
	TxID types.Hash
}

func (e EmptyOutputsError) Error() string {
	return fmt.Sprintf("отсутствуют выходы транзакции %v", e.TxID)
}

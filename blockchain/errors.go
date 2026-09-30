package blockchain

import (
	"fmt"
	"zxcoin/types"
)

type DoubleSpendError struct {
	TxID     types.Hash
	OutIndex uint64
}

func (e DoubleSpendError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) уже используется в данном блоке", e.TxID, e.OutIndex)
}

type InvalidNonceError struct {
	BlockHash  types.Hash
	Difficulty uint64
}

func (e InvalidNonceError) Error() string {
	return fmt.Sprintf("неверный nonce. Сложность: %v, хеш: %v", e.Difficulty, e.BlockHash)
}

type InvalidPrevHashError struct{}

func (e InvalidPrevHashError) Error() string {
	return "неверный предыдущий блок"
}

type MoreOneCoinbaseError struct{}

func (e MoreOneCoinbaseError) Error() string {
	return "больше одной coinbase-транзакции в блоке"
}

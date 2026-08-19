package blockchain

import "fmt"

type DoubleSpendError struct {
	TxID     [32]byte
	OutIndex int
}

func (e *DoubleSpendError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) уже используется в данном блоке", e.TxID, e.OutIndex)
}

type InvalidNonceError struct {
	BlockHash  [32]byte
	Difficulty int
}

func (e *InvalidNonceError) Error() string {
	return fmt.Sprintf("неверный nonce. Сложность: %v, хеш: %v", e.Difficulty, e.BlockHash)
}

type InvalidPrevHashError struct{}

func (e *InvalidPrevHashError) Error() string {
	return "неверный предыдущий блок"
}

type InvalidMerkleRootError struct{}

func (e *InvalidMerkleRootError) Error() string {
	return "неверный хеш корня дерева Меркла"
}

type MoreOneCoinbaseError struct{}

func (e *MoreOneCoinbaseError) Error() string {
	return "больше одной coinbase-транзакции в блоке"
}

package block

import (
	"fmt"

	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type InvalidMerkleRootError struct{}

func (e InvalidMerkleRootError) Error() string {
	return "неверный хеш корня дерева Меркла"
}

type InvalidNonceError struct {
	BlockHash  types.Hash
	Difficulty uint64
}

func (e InvalidNonceError) Error() string {
	return fmt.Sprintf("неверный nonce. Сложность: %v, хеш: %v", e.Difficulty, e.BlockHash)
}

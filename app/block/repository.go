package block

import (
	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type Repository interface {
	FindByHash(hash types.Hash) (block.Block, bool, error)
	FindByHeight(height uint64) ([]block.Block, error)
	List(from uint64, limit uint64) ([]block.Block, error)
	Save(block block.Block, height uint64) error
	Tip() (block.Block, uint64, bool, error)
}

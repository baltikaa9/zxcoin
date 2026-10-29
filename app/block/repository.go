package block

import (
	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type Repository interface {
	FindByHash(hash types.Hash) (*block.Block, error)
	FindByHeight(height int) (*block.Block, error)
	List(from int, limit int) ([]block.Block, error)
	Save(block block.Block, height int) error
	Tip() (*block.Block, int, error)
}

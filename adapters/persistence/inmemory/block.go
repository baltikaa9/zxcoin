package inmemory

import (
	"errors"
	"sync"

	appblock "github.com/baltikaa9/zxcoin/app/block"
	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

var errTransactionClosed = errors.New("transaction is closed")

var _ appblock.Repository = (*Block)(nil)

type Block struct {
	mu sync.RWMutex
	db map[utxo.ID]utxo.UTXO
}

func NewBlock() *Block {
	return &Block{db: make(map[utxo.ID]utxo.UTXO)}
}

func (br *Block) FindByHash(hash types.Hash) (block.Block, bool, error) {
	return block.Block{}, false, nil
}

func (br *Block) FindByHeight(height uint64) (block.Block, bool, error) {
	return block.Block{}, false, nil
}

func (br *Block) Tip() (block.Block, uint64, bool, error) {
	return block.Block{}, 0, false, nil
}

func (br *Block) List(from uint64, limit uint64) ([]block.Block, error) {
	return nil, nil
}

func (br *Block) Save(block block.Block, height uint64) error { return nil }

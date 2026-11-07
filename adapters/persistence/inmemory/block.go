package inmemory

import (
	"fmt"
	"slices"
	"sync"

	appblock "github.com/baltikaa9/zxcoin/app/block"
	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

var _ appblock.Repository = (*Block)(nil)

type Block struct {
	mu       sync.RWMutex
	blocks   []block.Block
	heights  map[types.Hash]uint64
	utxoRepo *UTXO
}

func NewBlock(utxoRepo *UTXO) *Block {
	return &Block{
		heights:  make(map[types.Hash]uint64),
		utxoRepo: utxoRepo,
	}
}

func (r *Block) FindByHash(hash types.Hash) (block.Block, uint64, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	height, exists := r.heights[hash]
	if !exists {
		return block.Block{}, 0, false, nil
	}

	return r.blocks[height], height, true, nil
}

func (r *Block) FindByHeight(height uint64) (block.Block, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if height >= uint64(len(r.blocks)) {
		return block.Block{}, false, nil
	}

	return r.blocks[height], true, nil
}

func (r *Block) Tip() (block.Block, uint64, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.blocks) == 0 {
		return block.Block{}, 0, false, nil
	}

	height := uint64(len(r.blocks) - 1)

	return r.blocks[height], height, true, nil
}

func (r *Block) List(from uint64, limit uint64) ([]block.Block, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := uint64(len(r.blocks))
	if from >= count {
		return []block.Block{}, nil
	}

	end := count
	if limit < count-from {
		end = from + limit
	}

	return slices.Clone(r.blocks[from:end]), nil
}

func (r *Block) Save(b block.Block, height uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if height != uint64(len(r.blocks)) {
		return fmt.Errorf("invalid block height: got %d, want %d", height, len(r.blocks))
	}

	inputs := make([]utxo.ID, 0)
	outputs := make([]utxo.UTXO, 0)

	for _, t := range b.Transactions {
		hash, err := t.Hash()
		if err != nil {
			return fmt.Errorf("hash transaction: %w", err)
		}

		for _, i := range t.Inputs {
			inputs = append(inputs, i.ID)
		}

		for i, o := range t.Outputs {
			outputs = append(outputs, utxo.UTXO{
				ID: utxo.ID{
					TxID:     hash,
					OutIndex: uint64(i),
				},
				Output: o,
			})
		}
	}

	for _, o := range outputs {
		r.utxoRepo.Save(o)
	}

	for _, i := range inputs {
		r.utxoRepo.Delete(i)
	}

	r.blocks = append(r.blocks, b)
	r.heights[b.Header.Hash()] = height

	return nil
}

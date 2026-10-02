// Package blockchain реализует основную логику консенсуса: проверку и добавление блоков, защиту от двойной траты и управление наградой за майнинг.
package blockchain

import (
	"github.com/baltikaa9/zxcoin/block"
)

type Blockchain struct {
	blocks            []block.Block
	currentDifficulty uint64
	currentAward      uint64
}

func NewBlockchain(difficulty uint64, award uint64) *Blockchain {
	return &Blockchain{currentDifficulty: difficulty, currentAward: award}
}

func (bc *Blockchain) AppendBlock(block block.Block) {
	bc.blocks = append(bc.blocks, block)
}

func (bc *Blockchain) VerifyPrevHash(block block.Block) error {
	if len(bc.blocks) > 0 && block.Header.PrevHash != bc.blocks[len(bc.blocks)-1].Header.Hash() {
		return InvalidPrevHashError{}
	}

	return nil
}

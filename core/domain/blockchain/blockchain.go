// Package blockchain реализует основную логику консенсуса: проверку и добавление блоков, защиту от двойной траты и управление наградой за майнинг.
package blockchain

type Blockchain struct {
	currentDifficulty uint64
	currentAward      uint64
}

func NewBlockchain(difficulty uint64, award uint64) *Blockchain {
	return &Blockchain{currentDifficulty: difficulty, currentAward: award}
}

func (bc *Blockchain) CurrentDifficulty() uint64 {
	return bc.currentDifficulty
}

func (bc *Blockchain) CurrentAward() uint64 {
	return bc.currentAward
}

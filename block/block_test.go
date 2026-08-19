package block

import (
	"testing"
	"zxcoin/transaction"
)

func TestMine(t *testing.T) {
	b := Block{
		Header:       BlockHeader{PrevHash: [32]byte{}, RootHash: [32]byte{}, Timestamp: 0},
		Transactions: []transaction.Transaction{},
		Difficulty:   2,
	}

	b.Mine()

	hash := b.Header.Hash()

	for i := range b.Difficulty {
		if hash[i] != 0 {
			t.Fatalf("хеш не удовлетворяет сложности %d: %v", b.Difficulty, hash)
		}
	}
}

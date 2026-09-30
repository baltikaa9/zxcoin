package block

import (
	"testing"
	"zxcoin/transaction"
)

func TestMine(t *testing.T) {
	difficulty := uint64(2)
	b := Block{
		Header:       BlockHeader{PrevHash: [32]byte{}, RootHash: [32]byte{}, Timestamp: 0},
		Transactions: []transaction.Transaction{},
	}
	b.Mine(difficulty)
	hash := b.Header.Hash()

	for i := range difficulty {
		if hash[i] != 0 {
			t.Fatalf("хеш не удовлетворяет сложности %d: %v", difficulty, hash)
		}
	}
}

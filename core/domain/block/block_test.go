package block

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
)

func TestMine(t *testing.T) {
	difficulty := uint64(2)
	b := Block{
		Header:       Header{PrevHash: types.Hash{}, RootHash: types.Hash{}, Timestamp: 0},
		Transactions: []transaction.Transaction{},
	}
	_ = b.Mine(t.Context(), difficulty)
	hash := b.Header.Hash()

	for i := range difficulty {
		if hash[i] != 0 {
			t.Fatalf("хеш не удовлетворяет сложности %d: %v", difficulty, hash)
		}
	}
}

func TestMine_Cancel(t *testing.T) {
	testError := errors.New("test")

	difficulty := uint64(42)
	b := Block{
		Header:       Header{PrevHash: types.Hash{}, RootHash: types.Hash{}, Timestamp: 0},
		Transactions: []transaction.Transaction{},
	}

	ctx, cancel := context.WithCancelCause(t.Context())
	defer cancel(nil)

	done := make(chan error)

	go func() {
		done <- b.Mine(ctx, difficulty)
	}()

	cancel(testError)

	select {
	case err := <-done:
		if !errors.Is(err, testError) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("Mine не остановился после отмены")
	}

	if b.Header.Nonce != 0 {
		t.Fatal("nonce не обнулился после отмены")
	}
}

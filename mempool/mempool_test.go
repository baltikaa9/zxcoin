package mempool

import (
	"testing"

	"github.com/baltikaa9/zxcoin/coin"
	"github.com/baltikaa9/zxcoin/testutil"
	"github.com/baltikaa9/zxcoin/transaction"
	"github.com/baltikaa9/zxcoin/types"
	"github.com/baltikaa9/zxcoin/utxo"
	"github.com/baltikaa9/zxcoin/utxo/inmemory"
)

func TestAdd(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := transaction.NewValidator(repo)
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)
	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: amount, Owner: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, hash)
	mempool := NewMempool(validator)
	err = mempool.Add(tx)

	if err != nil {
		t.Fatalf("ошибка при добавлении транзакции: %v", err)
	}

	_, ok := mempool.transactions[hash]

	if !ok {
		t.Fatalf("транзакция не добавилась")
	}
}

func TestRemove(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: 42, Owner: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	mempool := Mempool{transactions: map[types.Hash]transaction.Transaction{hash: tx}}
	mempool.Remove(hash)
	_, ok := mempool.transactions[hash]

	if ok {
		t.Fatalf("транзакция не удалилась")
	}
}

func TestGetPending_LessThanLimit(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	tx1 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: 42, Owner: publicKey}},
	}
	tx1Hash, err := tx1.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{OutIndex: 1}}},
		Outputs: []coin.TxOutput{{Amount: 42, Owner: publicKey}},
	}
	tx2Hash, err := tx2.Hash()

	if err != nil {
		t.Fatal(err)
	}

	txMap := map[types.Hash]transaction.Transaction{
		tx1Hash: tx1,
		tx2Hash: tx2,
	}
	mempool := Mempool{transactions: txMap}
	txs := mempool.GetPending(3)

	if len(txs) != len(txMap) {
		t.Fatalf("неверное количество транзакций: ожидалось %v, вернулось %v", len(txMap), len(txs))
	}
}

func TestGetPending_MoreThanLimit(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	tx1 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: 42, Owner: publicKey}},
	}
	tx1Hash, err := tx1.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{OutIndex: 1}}},
		Outputs: []coin.TxOutput{{Amount: 42, Owner: publicKey}},
	}
	tx2Hash, err := tx2.Hash()

	if err != nil {
		t.Fatal(err)
	}

	mempool := Mempool{transactions: map[types.Hash]transaction.Transaction{
		tx1Hash: tx1,
		tx2Hash: tx2,
	}}
	limit := uint64(1)
	txs := mempool.GetPending(limit)

	if len(txs) != int(limit) {
		t.Fatalf("неверное количество транзакций: ожидалось %v, вернулось %v", limit, len(txs))
	}
}

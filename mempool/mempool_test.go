package mempool

import (
	"testing"
	"zxcoin/coin"
	"zxcoin/testutil"
	"zxcoin/transaction"
)

func TestAdd(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := uint64(5)
	utxoDB := testutil.GenerateSingleUtxo(t, amount, publicKey)
	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: amount, PublicKey: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, hash)
	mempool := NewMempool()
	err = mempool.Add(tx, utxoDB)

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
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 42, PublicKey: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	mempool := Mempool{transactions: map[[32]byte]transaction.Transaction{hash: tx}}
	mempool.Remove(hash)
	_, ok := mempool.transactions[hash]

	if ok {
		t.Fatalf("транзакция не удалилась")
	}
}

func TestGetPending_LessThanLimit(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	tx1 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 42, PublicKey: publicKey}},
	}
	tx1Hash, err := tx1.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 1}},
		Outputs: []coin.TxOutput{{Amount: 42, PublicKey: publicKey}},
	}
	tx2Hash, err := tx2.Hash()

	if err != nil {
		t.Fatal(err)
	}

	txMap := map[[32]byte]transaction.Transaction{
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
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 42, PublicKey: publicKey}},
	}
	tx1Hash, err := tx1.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 1}},
		Outputs: []coin.TxOutput{{Amount: 42, PublicKey: publicKey}},
	}
	tx2Hash, err := tx2.Hash()

	if err != nil {
		t.Fatal(err)
	}

	mempool := Mempool{transactions: map[[32]byte]transaction.Transaction{
		tx1Hash: tx1,
		tx2Hash: tx2,
	}}
	limit := 1
	txs := mempool.GetPending(limit)

	if len(txs) != limit {
		t.Fatalf("неверное количество транзакций: ожидалось %v, вернулось %v", limit, len(txs))
	}
}

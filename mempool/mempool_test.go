package mempool

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"zxcoin/coin"
	"zxcoin/transaction"
	"zxcoin/utxo"
)

func TestAdd(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey
	amount := 5
	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    amount,
				PublicKey: publicKey,
			},
		},
	}
	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: amount}},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())
	mempool := NewMempool()
	err := mempool.Add(tx, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при добавлении транзакции: %v", err)
	}

	_, ok := mempool.transactions[tx.Hash()]

	if !ok {
		t.Fatalf("транзакция не добавилась")
	}
}

func TestRemove(t *testing.T) {
	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 42}},
	}
	mempool := Mempool{transactions: map[[32]byte]transaction.Transaction{tx.Hash(): tx}}
	mempool.Remove(tx.Hash())
	_, ok := mempool.transactions[tx.Hash()]

	if ok {
		t.Fatalf("транзакция не удалилась")
	}
}

func TestGetPending_LessThanLimit(t *testing.T) {
	tx1 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 42}},
	}
	tx2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 1}},
		Outputs: []coin.TxOutput{{Amount: 42}},
	}
	txMap := map[[32]byte]transaction.Transaction{
		tx1.Hash(): tx1,
		tx2.Hash(): tx2,
	}
	mempool := Mempool{transactions: txMap}
	txs := mempool.GetPending(3)

	if len(txs) != len(txMap) {
		t.Fatalf("неверное количество транзакций: ожидалось %v, вернулось %v", len(txMap), len(txs))
	}
}

func TestGetPending_MoreThanLimit(t *testing.T) {
	tx1 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 42}},
	}
	tx2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 1}},
		Outputs: []coin.TxOutput{{Amount: 42}},
	}
	mempool := Mempool{transactions: map[[32]byte]transaction.Transaction{
		tx1.Hash(): tx1,
		tx2.Hash(): tx2,
	}}
	limit := 1
	txs := mempool.GetPending(limit)

	if len(txs) != limit {
		t.Fatalf("неверное количество транзакций: ожидалось %v, вернулось %v", limit, len(txs))
	}
}

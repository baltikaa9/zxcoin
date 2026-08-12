package main

import (
	"errors"
	"testing"
)

func TestAddBlock_DoubleSpendInBlock(t *testing.T) {
	wallet := NewWallet()
	other := NewWallet()

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: UTXOEntry{TxOutput{5, wallet.PublicKey}, false},
	}
	mempool := NewMempool()

	t1, err := wallet.CreateTransaction(other.PublicKey, 5, utxoDB)

	if err != nil {
		t.Errorf("ошибка при создании честной транзакции: %v", err)
	}

	t2 := Transaction{[]TxInput{{TxID: [32]byte{}, OutIndex: 0}}, []TxOutput{{5, wallet.PublicKey}}}
	t2.Inputs[0].Sign(wallet.PrivateKey, t2.Hash())

	err = mempool.Add(t1, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции в пул: %v", err)
	}

	err = mempool.Add(t2, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции в пул: %v", err)
	}

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	_, err = bc.MineAndAddBlock(mempool, utxoDB, 2, wallet.PublicKey)

	if err == nil {
		t.Fatalf("ожидалась ошибка двойной траты, но AddBlock прошёл успешно")
	}

	var dsErr *DoubleSpendError

	if !errors.As(err, &dsErr) {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_DoubleSpendInTransaction(t *testing.T) {
	wallet := NewWallet()
	other := NewWallet()

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: UTXOEntry{TxOutput{5, wallet.PublicKey}, false},
	}
	mempool := NewMempool()

	t1 := Transaction{
		Inputs: []TxInput{
			{TxID: [32]byte{}, OutIndex: 0},
			{TxID: [32]byte{}, OutIndex: 0},
		},
		Outputs: []TxOutput{
			{5, other.PublicKey},
		},
	}

	hash := t1.Hash()

	for i := range t1.Inputs {
		t1.Inputs[i].Sign(wallet.PrivateKey, hash)
	}

	err := mempool.Add(t1, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции: %v", err)
	}

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	_, err = bc.MineAndAddBlock(mempool, utxoDB, 1, wallet.PublicKey)

	if err == nil {
		t.Fatalf("ожидалась ошибка двойной траты, но AddBlock прошёл успешно")
	}

	var dsErr *DoubleSpendError

	if !errors.As(err, &dsErr) {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_DoubleSpendInBlockchain(t *testing.T) {
	wallet := NewWallet()
	other := NewWallet()

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: UTXOEntry{TxOutput{5, wallet.PublicKey}, false},
	}
	mempool := NewMempool()

	t1, err := wallet.CreateTransaction(other.PublicKey, 5, utxoDB)

	if err != nil {
		t.Errorf("ошибка при создании честной транзакции: %v", err)
	}

	err = mempool.Add(t1, utxoDB)

	// fmt.Printf("%v", err)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции: %v", err)
	}

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	_, err = bc.MineAndAddBlock(mempool, utxoDB, 1, wallet.PublicKey)

	if err != nil {
		t.Fatalf("%v", err)
	}

	t2 := Transaction{[]TxInput{{TxID: [32]byte{}, OutIndex: 0}}, []TxOutput{{5, wallet.PublicKey}}}
	t2.Inputs[0].Sign(wallet.PrivateKey, t2.Hash())

	err = mempool.Add(t2, utxoDB)

	var nfErr *UTXONotFoundError

	if !errors.As(err, &nfErr) {
		t.Fatalf("ожидалась UTXONotFoundError, получено: %v", err)
	}
}

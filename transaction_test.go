package main

import (
	"errors"
	"testing"
)

func TestValidate_UTXONotFound(t *testing.T) {
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

	if _, ok := errors.AsType[*UTXONotFoundError](err); !ok {
		t.Fatalf("ожидалась UTXONotFoundError, получено: %v", err)
	}
}

func TestValidate_InvalidSignature(t *testing.T) {
	wallet := NewWallet()
	other := NewWallet()
	attacker := NewWallet()

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: UTXOEntry{TxOutput{5, wallet.PublicKey}, false},
	}
	mempool := NewMempool()

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []TxOutput{{Amount: 5, PublicKey: other.PublicKey}},
	}
	tx.Inputs[0].Sign(attacker.PrivateKey, tx.Hash())

	err := mempool.Add(tx, utxoDB)

	if _, ok := errors.AsType[*InvalidSignatureError](err); !ok {
		t.Fatalf("ожидалась InvalidSignatureError, получено: %v", err)
	}
}

func TestValidate_NotEnoughMoney(t *testing.T) {
	wallet := NewWallet()
	other := NewWallet()

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: UTXOEntry{TxOutput{5, wallet.PublicKey}, false},
	}
	mempool := NewMempool()

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []TxOutput{{Amount: 10, PublicKey: other.PublicKey}},
	}
	tx.Inputs[0].Sign(wallet.PrivateKey, tx.Hash())

	err := mempool.Add(tx, utxoDB)

	if _, ok := errors.AsType[*NotEnoughMoneyError](err); !ok {
		t.Fatalf("ожидалась NotEnoughMoney, получено: %v", err)
	}
}

func TestValidate_Success(t *testing.T) {
	wallet := NewWallet()
	other := NewWallet()

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: UTXOEntry{TxOutput{5, wallet.PublicKey}, false},
	}
	mempool := NewMempool()

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []TxOutput{{Amount: 5, PublicKey: other.PublicKey}},
	}
	tx.Inputs[0].Sign(wallet.PrivateKey, tx.Hash())

	err := mempool.Add(tx, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при валидации: %v", err)
	}
}
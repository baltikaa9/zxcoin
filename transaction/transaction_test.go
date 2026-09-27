package transaction

import (
	"errors"
	"testing"
	"zxcoin/coin"
	"zxcoin/testutil"
	"zxcoin/utxo"
	"zxcoin/utxo/inmemory"
)

func TestValidate_UTXONotFound(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()

	tx := Transaction{
		Inputs:  []TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, hash)

	err = tx.Validate(repo)

	if _, ok := errors.AsType[*UTXONotFoundError](err); !ok {
		t.Fatalf("ожидалась UTXONotFoundError, получено: %v", err)
	}
}

func TestValidate_InvalidSignature(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	attackerPrivateKey, _ := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := Transaction{
		Inputs:  []TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: amount, PublicKey: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(attackerPrivateKey, hash)

	err = tx.Validate(repo)

	if _, ok := errors.AsType[*InvalidSignatureError](err); !ok {
		t.Fatalf("ожидалась InvalidSignatureError, получено: %v", err)
	}
}

func TestValidate_InsufficientFunds(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := Transaction{
		Inputs:  []TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: amount * 2, PublicKey: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, hash)

	err = tx.Validate(repo)

	if _, ok := errors.AsType[*InsufficientFundsError](err); !ok {
		t.Fatalf("ожидалась InsufficientFundsError, получено: %v", err)
	}
}

func TestValidate_NonPositiveOutputZero(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := Transaction{
		Inputs:  []TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: 0, PublicKey: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, hash)

	err = tx.Validate(repo)

	if _, ok := errors.AsType[*ZeroOutputError](err); !ok {
		t.Fatalf("ожидалась ZeroOutputError, получено: %v", err)
	}
}

func TestValidate_Success(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := Transaction{
		Inputs:  []TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: amount, PublicKey: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, hash)

	err = tx.Validate(repo)

	if err != nil {
		t.Fatalf("ошибка при валидации: %v", err)
	}
}

package transaction

import (
	"errors"
	"testing"

	"github.com/baltikaa9/zxcoin/adapters/persistence/inmemory"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
	"github.com/baltikaa9/zxcoin/core/testutil"
)

func TestValidate_UTXONotFound(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := TransactionValidator{repo}

	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: 5, Owner: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, hash)

	err = validator.Validate(tx)

	if _, ok := errors.AsType[UTXONotFoundError](err); !ok {
		t.Fatalf("ожидалась UTXONotFoundError, получено: %v", err)
	}
}

func TestValidate_InvalidSignature(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	attackerPrivateKey, _ := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := TransactionValidator{repo}
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

	tx.Inputs[0].Sign(attackerPrivateKey, hash)

	err = validator.Validate(tx)

	if _, ok := errors.AsType[transaction.InvalidSignatureError](err); !ok {
		t.Fatalf("ожидалась InvalidSignatureError, получено: %v", err)
	}
}

func TestValidate_InsufficientFunds(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := TransactionValidator{repo}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: amount * 2, Owner: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, hash)

	err = validator.Validate(tx)

	if _, ok := errors.AsType[transaction.InsufficientFundsError](err); !ok {
		t.Fatalf("ожидалась InsufficientFundsError, получено: %v", err)
	}
}

func TestValidate_NonPositiveOutputZero(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := TransactionValidator{repo}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: 0, Owner: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, hash)

	err = validator.Validate(tx)

	if _, ok := errors.AsType[transaction.ZeroOutputError](err); !ok {
		t.Fatalf("ожидалась ZeroOutputError, получено: %v", err)
	}
}

func TestValidate_Success(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := TransactionValidator{repo}
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

	err = validator.Validate(tx)

	if err != nil {
		t.Fatalf("ошибка при валидации: %v", err)
	}
}

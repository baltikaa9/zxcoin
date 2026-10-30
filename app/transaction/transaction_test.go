package transaction

import (
	"testing"

	"github.com/baltikaa9/zxcoin/adapters/persistence/inmemory"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
	"github.com/baltikaa9/zxcoin/core/testutil"
)

func TestValidate_UTXONotFound(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewBlock()
	validator := Validator{repo}

	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.ID{}}},
		Outputs: []coin.TxOutput{{Amount: 5, Owner: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Inputs[0].Sign(privateKey, hash); err != nil {
		t.Fatal(err)
	}

	err = validator.Validate(tx)

	testutil.CheckError[UTXONotFoundError](t, err)
}

func TestValidate_InvalidSignature(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	attackerPrivateKey, _ := testutil.GenerateKeyPair(t)
	repo := inmemory.NewBlock()
	validator := Validator{repo}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.ID{}}},
		Outputs: []coin.TxOutput{{Amount: amount, Owner: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Inputs[0].Sign(attackerPrivateKey, hash); err != nil {
		t.Fatal(err)
	}

	err = validator.Validate(tx)

	testutil.CheckError[transaction.InvalidSignatureError](t, err)
}

func TestValidate_InsufficientFunds(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewBlock()
	validator := Validator{repo}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.ID{}}},
		Outputs: []coin.TxOutput{{Amount: amount * 2, Owner: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Inputs[0].Sign(privateKey, hash); err != nil {
		t.Fatal(err)
	}

	err = validator.Validate(tx)

	testutil.CheckError[transaction.InsufficientFundsError](t, err)
}

func TestValidate_NonPositiveOutputZero(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewBlock()
	validator := Validator{repo}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.ID{}}},
		Outputs: []coin.TxOutput{{Amount: 0, Owner: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Inputs[0].Sign(privateKey, hash); err != nil {
		t.Fatal(err)
	}

	err = validator.Validate(tx)

	testutil.CheckError[transaction.ZeroOutputError](t, err)
}

func TestValidate_Success(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewBlock()
	validator := Validator{repo}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.ID{}}},
		Outputs: []coin.TxOutput{{Amount: amount, Owner: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Inputs[0].Sign(privateKey, hash); err != nil {
		t.Fatal(err)
	}

	err = validator.Validate(tx)

	if err != nil {
		t.Fatalf("ошибка при валидации: %v", err)
	}
}

func TestValidate_DuplicateInputs(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewBlock()
	validator := Validator{repo}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.ID{}}, {ID: utxo.ID{}}},
		Outputs: []coin.TxOutput{{Amount: amount, Owner: publicKey}},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Inputs[0].Sign(privateKey, hash); err != nil {
		t.Fatal(err)
	}

	err = validator.Validate(tx)

	testutil.CheckError[transaction.DuplicateInputError](t, err)
}

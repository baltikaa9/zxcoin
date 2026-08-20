package transaction

import (
	"errors"
	"testing"
	"zxcoin/coin"
	"zxcoin/testutil"
	"zxcoin/utxo"
)

func TestValidate_UTXONotFound(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)

	utxoDB := utxo.UTXODB{}

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: publicKey}},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())

	err := tx.Validate(utxoDB)

	if _, ok := errors.AsType[*UTXONotFoundError](err); !ok {
		t.Fatalf("ожидалась UTXONotFoundError, получено: %v", err)
	}
}

func TestValidate_InvalidSignature(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	attackerPrivateKey, _ := testutil.GenerateKeyPair(t)
	amount := 5
	utxoDB := testutil.GenerateSingleUtxo(t, amount, publicKey)

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: amount, PublicKey: publicKey}},
	}
	tx.Inputs[0].Sign(attackerPrivateKey, tx.Hash())

	err := tx.Validate(utxoDB)

	if _, ok := errors.AsType[*InvalidSignatureError](err); !ok {
		t.Fatalf("ожидалась InvalidSignatureError, получено: %v", err)
	}
}

func TestValidate_InsufficientFunds(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := 5
	utxoDB := testutil.GenerateSingleUtxo(t, amount, publicKey)

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: amount * 2, PublicKey: publicKey}},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())

	err := tx.Validate(utxoDB)

	if _, ok := errors.AsType[*InsufficientFundsError](err); !ok {
		t.Fatalf("ожидалась InsufficientFundsError, получено: %v", err)
	}
}

func TestValidate_NonPositiveOutputZero(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := 5
	utxoDB := testutil.GenerateSingleUtxo(t, amount, publicKey)

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 0, PublicKey: publicKey}},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())

	err := tx.Validate(utxoDB)

	if _, ok := errors.AsType[*NonPositiveOutputError](err); !ok {
		t.Fatalf("ожидалась NonPositiveOutputError, получено: %v", err)
	}
}

func TestValidate_NonPositiveOutputNegative(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := 5
	utxoDB := testutil.GenerateSingleUtxo(t, amount, publicKey)

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: -amount, PublicKey: publicKey}},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())

	err := tx.Validate(utxoDB)

	if _, ok := errors.AsType[*NonPositiveOutputError](err); !ok {
		t.Fatalf("ожидалась NonPositiveOutputError, получено: %v", err)
	}
}

func TestValidate_Success(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := 5
	utxoDB := testutil.GenerateSingleUtxo(t, amount, publicKey)

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: amount, PublicKey: publicKey}},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())

	err := tx.Validate(utxoDB)

	if err != nil {
		t.Fatalf("ошибка при валидации: %v", err)
	}
}

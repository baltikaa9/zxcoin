package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"testing"
	"zxcoin/coin"
	"zxcoin/utxo"
)

func TestValidate_UTXONotFound(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

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
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	attackerPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: publicKey,
			},
		},
	}

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: publicKey}},
	}
	tx.Inputs[0].Sign(attackerPrivateKey, tx.Hash())

	err := tx.Validate(utxoDB)

	if _, ok := errors.AsType[*InvalidSignatureError](err); !ok {
		t.Fatalf("ожидалась InvalidSignatureError, получено: %v", err)
	}
}

func TestValidate_NotEnoughMoney(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: publicKey,
			},
		},
	}

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 10, PublicKey: publicKey}},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())

	err := tx.Validate(utxoDB)

	if _, ok := errors.AsType[*InsufficientFundsError](err); !ok {
		t.Fatalf("ожидалась InsufficientFundsError, получено: %v", err)
	}
}

func TestValidate_NonPositiveOutputZero(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: publicKey,
			},
		},
	}

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
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: publicKey,
			},
		},
	}

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: -10, PublicKey: publicKey}},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())

	err := tx.Validate(utxoDB)

	if _, ok := errors.AsType[*NonPositiveOutputError](err); !ok {
		t.Fatalf("ожидалась NonPositiveOutputError, получено: %v", err)
	}
}

func TestValidate_Success(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: publicKey,
			},
		},
	}

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: publicKey}},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())

	err := tx.Validate(utxoDB)

	if err != nil {
		t.Fatalf("ошибка при валидации: %v", err)
	}
}

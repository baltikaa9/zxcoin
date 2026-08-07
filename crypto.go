package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

func NewWallet() Wallet {
	privateKey, err := ecdsa.GenerateKey(
		elliptic.P256(),
		rand.Reader,
	)
	if err != nil {
		panic(err)
	}

	publicKey := &privateKey.PublicKey

	return Wallet{privateKey, publicKey}
}

func (w Wallet) CreateTransaction(to *ecdsa.PublicKey, amount int, utxoDB UTXODB) (Transaction, error) {
	var inputs []TxInput
	var outputs []TxOutput

	total := 0

	for key, output := range utxoDB {
		if output.PublicKey.Equal(w.PublicKey) {
			inputs = append(inputs, TxInput{
				TxID:     key.TxID,
				OutIndex: key.OutIndex,
			})

			total += output.Amount

			if total >= amount {
				break
			}
		}
	}

	if total < amount {
		return Transaction{}, fmt.Errorf("недостаточно средств %v/%v", total, amount)
	}

	outputs = append(outputs, TxOutput{amount, to})

	change := total - amount

	if change > 0 {
		outputs = append(outputs, TxOutput{change, w.PublicKey})
	}

	t := NewTransaction(inputs, outputs)
	hash := t.Hash()

	for i := range t.Inputs {
		t.Inputs[i].Sign(w.PrivateKey, hash)
	}

	return t, nil
}

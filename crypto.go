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
	var reserved []UTXOKey

	total := 0

	for key, entry := range utxoDB {
		if entry.Output.PublicKey.Equal(w.PublicKey) && !entry.Reserved {
			inputs = append(inputs, TxInput{
				TxID:     key.TxID,
				OutIndex: key.OutIndex,
			})

			reserved = append(reserved, key)

			total += entry.Output.Amount

			if total >= amount {
				break
			}
		}
	}

	if total < amount {
		return Transaction{}, fmt.Errorf("недостаточно средств %v/%v", total, amount)
	}

	for _, key := range reserved {
		item := utxoDB[key]
		item.Reserved = true
		utxoDB[key] = item
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

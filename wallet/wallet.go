package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"zxcoin/coin"
	"zxcoin/transaction"
	"zxcoin/utxo"
)

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

func (w Wallet) CreateTransaction(to *ecdsa.PublicKey, amount int, utxoDB utxo.UTXODB) (transaction.Transaction, error) {
	var inputs []transaction.TxInput
	var outputs []coin.TxOutput
	var reserved []utxo.UTXOKey

	total := 0

	for key, entry := range utxoDB {
		if entry.Output.PublicKey.Equal(w.PublicKey) && !entry.Reserved {
			inputs = append(inputs, transaction.TxInput{
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
		return transaction.Transaction{}, &InsufficientFundsError{Available: total, Requested: amount}
	}

	for _, key := range reserved {
		item := utxoDB[key]
		item.Reserved = true
		utxoDB[key] = item
	}

	outputs = append(outputs, coin.TxOutput{Amount: amount, PublicKey: to})

	change := total - amount

	if change > 0 {
		outputs = append(outputs, coin.TxOutput{Amount: change, PublicKey: w.PublicKey})
	}

	t := transaction.Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
	hash := t.Hash()

	for i := range t.Inputs {
		t.Inputs[i].Sign(w.PrivateKey, hash)
	}

	return t, nil
}

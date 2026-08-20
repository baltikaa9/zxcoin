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
	inputs, total, err := w.selectInputs(amount, utxoDB)

	if err != nil {
		return transaction.Transaction{}, err
	}

	if err := w.reserveInputs(inputs, utxoDB); err != nil {
		return transaction.Transaction{}, err
	}

	t := transaction.NewTransaction(inputs, w.createOutputs(to, amount, total), w.PrivateKey)

	return t, nil
}

func (w Wallet) selectInputs(amount int, utxoDB utxo.UTXODB) ([]transaction.TxInput, int, error) {
	inputs := make([]transaction.TxInput, 0)
	total := 0

	for key, entry := range utxoDB {
		if entry.Output.PublicKey.Equal(w.PublicKey) && !entry.Reserved() {
			inputs = append(inputs, transaction.TxInput{
				TxID:     key.TxID,
				OutIndex: key.OutIndex,
			})
			total += entry.Output.Amount

			if total >= amount {
				break
			}
		}
	}

	if total < amount {
		return []transaction.TxInput{}, 0, &InsufficientFundsError{Available: total, Requested: amount}
	}

	return inputs, total, nil
}

func (w Wallet) reserveInputs(inputs []transaction.TxInput, utxoDB utxo.UTXODB) error {
	reserved := make([]utxo.UTXOKey, 0, len(inputs))

	for _, input := range inputs {
		key := utxo.UTXOKey{
			TxID:     input.TxID,
			OutIndex: input.OutIndex,
		}

		if err := utxoDB.Reserve(key); err != nil {
			for _, key := range reserved {
				utxoDB.Release(key)
			}

			return err
		}

		reserved = append(reserved, key)
	}

	return nil
}

func (w Wallet) createOutputs(to *ecdsa.PublicKey, amount int, total int) []coin.TxOutput {
	outputs := []coin.TxOutput{{Amount: amount, PublicKey: to}}
	change := total - amount

	if change > 0 {
		outputs = append(outputs, coin.TxOutput{Amount: change, PublicKey: w.PublicKey})
	}

	return outputs
}

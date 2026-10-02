// Package wallet предоставляет кошелёк — пару криптографических ключей — и логику создания подписанных транзакций на основе доступных UTXO.
package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/baltikaa9/zxcoin/coin"
	"github.com/baltikaa9/zxcoin/transaction"
	"github.com/baltikaa9/zxcoin/utxo"
)

type Wallet struct {
	PrivateKey         *ecdsa.PrivateKey
	PublicKey          *ecdsa.PublicKey
	reservationTracker *ReservationTracker
	utxoRepo           utxo.Repository
}

func NewWallet(tracker *ReservationTracker, utxoRepo utxo.Repository) Wallet {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// Ошибка генерации ключа означает неисправность криптографической подсистемы,
	// при которой продолжение работы невозможно.
	if err != nil {
		panic(err)
	}

	return Wallet{
		PrivateKey:         privateKey,
		PublicKey:          &privateKey.PublicKey,
		reservationTracker: tracker,
		utxoRepo:           utxoRepo,
	}
}

func LoadWallet(privateKey *ecdsa.PrivateKey, tracker *ReservationTracker, utxoRepo utxo.Repository) Wallet {
	return Wallet{
		PrivateKey:         privateKey,
		PublicKey:          &privateKey.PublicKey,
		reservationTracker: tracker,
		utxoRepo:           utxoRepo,
	}
}

func (w Wallet) CreateTransaction(to *ecdsa.PublicKey, amount uint64) (transaction.Transaction, error) {
	inputs, total, err := w.selectInputs(amount)

	if err != nil {
		return transaction.Transaction{}, err
	}

	t, err := transaction.NewTransaction(inputs, w.createOutputs(to, amount, total), w.PrivateKey)

	if err != nil {
		return transaction.Transaction{}, err
	}

	w.reserveInputs(inputs)

	return t, nil
}

func (w Wallet) selectInputs(amount uint64) ([]transaction.TxInput, uint64, error) {
	inputs := make([]transaction.TxInput, 0)
	total := uint64(0)

	utxos, err := w.utxoRepo.FindByOwner(w.PublicKey)

	if err != nil {
		return []transaction.TxInput{}, 0, err
	}

	for _, utxo := range utxos {
		if reserved := w.reservationTracker.IsReserved(utxo.ID); !reserved {
			inputs = append(inputs, transaction.TxInput{ID: utxo.ID})
			total += utxo.Output.Amount

			if total >= amount {
				break
			}
		}
	}

	if total < amount {
		return []transaction.TxInput{}, 0, InsufficientFundsError{Available: total, Requested: amount}
	}

	return inputs, total, nil
}

func (w Wallet) reserveInputs(inputs []transaction.TxInput) {
	for _, input := range inputs {
		w.reservationTracker.Reserve(input.ID)
	}
}

func (w Wallet) createOutputs(to *ecdsa.PublicKey, amount uint64, total uint64) []coin.TxOutput {
	outputs := []coin.TxOutput{{Amount: amount, Owner: to}}
	change := total - amount

	if change > 0 {
		outputs = append(outputs, coin.TxOutput{Amount: change, Owner: w.PublicKey})
	}

	return outputs
}

func SavePrivateKey(path string, key *ecdsa.PrivateKey) error {
	der, err := x509.MarshalECPrivateKey(key)

	if err != nil {
		return err
	}

	data := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	})

	return os.WriteFile(path, data, 0600)
}

func LoadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)

	if block == nil {
		return nil, fmt.Errorf("invalid PEM private key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	return key, nil
}

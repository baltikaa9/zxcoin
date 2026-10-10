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

	"github.com/baltikaa9/zxcoin/app/logger"
	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type Wallet struct {
	PrivateKey         *ecdsa.PrivateKey
	PublicKey          types.PublicKey
	reservationTracker *ReservationTracker
	utxoRepo           apputxo.Repository
	logger             logger.Logger
}

func NewWallet(
	tracker *ReservationTracker,
	utxoRepo apputxo.Repository,
	logger logger.Logger,
) Wallet {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// Ошибка генерации ключа означает неисправность криптографической подсистемы,
	// при которой продолжение работы невозможно.
	if err != nil {
		panic(err)
	}

	return Wallet{
		PrivateKey:         privateKey,
		PublicKey:          types.PublicKey{PublicKey: &privateKey.PublicKey},
		reservationTracker: tracker,
		utxoRepo:           utxoRepo,
		logger:             logger.Named("wallet"),
	}
}

func LoadWallet(
	privateKey *ecdsa.PrivateKey,
	tracker *ReservationTracker,
	utxoRepo apputxo.Repository,
	logger logger.Logger,
) Wallet {
	return Wallet{
		PrivateKey:         privateKey,
		PublicKey:          types.PublicKey{PublicKey: &privateKey.PublicKey},
		reservationTracker: tracker,
		utxoRepo:           utxoRepo,
		logger:             logger.Named("wallet"),
	}
}

func (w Wallet) CreateTransaction(to types.PublicKey, amount uint64) (transaction.Transaction, error) {
	w.logger.Debug("creating transaction", "to", to.Short(), "amount", amount)

	inputs, total, err := w.selectInputs(amount)

	if err != nil {
		return transaction.Transaction{}, fmt.Errorf("select inputs: %w", err)
	}

	t, err := transaction.NewTransaction(inputs, w.createOutputs(to, amount, total), w.PrivateKey)

	if err != nil {
		return transaction.Transaction{}, fmt.Errorf("new transaction: %w", err)
	}

	hash, _ := t.Hash()

	w.reserveInputs(inputs)
	w.logger.Info(
		"transaction created",
		"tx_id", hash.String(),
		"to", to.Short(),
		"inputs", len(inputs),
		"amount", amount,
	)

	return t, nil
}

func (w Wallet) selectInputs(amount uint64) ([]transaction.TxInput, uint64, error) {
	w.logger.Debug("selecting inputs", "amount", amount)
	inputs := make([]transaction.TxInput, 0)
	total := uint64(0)

	utxos, err := w.utxoRepo.FindByOwner(w.PublicKey)

	if err != nil {
		return []transaction.TxInput{}, 0, fmt.Errorf("find by owner: %w", err)
	}

	w.logger.Debug("found by owner", "utxos", utxos)

	for _, utxo := range utxos {
		if reserved := w.reservationTracker.IsReserved(utxo.ID); !reserved {
			inputs = append(inputs, transaction.TxInput{ID: utxo.ID})
			total += utxo.Output.Amount
			w.logger.Info("selected utxo", "tx_id", utxo.ID.TxID.String(), "out_index", utxo.ID.OutIndex)

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
	w.logger.Debug("reserving inputs")

	for _, input := range inputs {
		w.reservationTracker.Reserve(input.ID)
		w.logger.Info("reserved utxo", "tx_id", input.ID.TxID.String(), "out_index", input.ID.OutIndex)
	}
}

func (w Wallet) createOutputs(to types.PublicKey, amount uint64, total uint64) []coin.TxOutput {
	w.logger.Debug("creating outputs", "amount", amount, "owner", to.Short())
	outputs := []coin.TxOutput{{Amount: amount, Owner: to}}
	w.logger.Info("created output", "amount", amount, "owner", to.Short())
	change := total - amount

	if change > 0 {
		outputs = append(outputs, coin.TxOutput{Amount: change, Owner: w.PublicKey})
		w.logger.Info("created change", "amount", change, "owner", w.PublicKey.Short())
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

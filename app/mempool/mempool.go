// Package mempool хранит транзакции, ожидающие включения в блок, и проверяет их валидность перед добавлением в пул.
package mempool

import (
	"github.com/baltikaa9/zxcoin/app/logger"
	apptransaction "github.com/baltikaa9/zxcoin/app/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type Mempool struct {
	transactions map[types.Hash]transaction.Transaction
	validator    *apptransaction.TransactionValidator
	logger       logger.Logger
}

func NewMempool(validator *apptransaction.TransactionValidator, logger logger.Logger) *Mempool {
	return &Mempool{
		transactions: make(map[types.Hash]transaction.Transaction),
		validator:    validator,
		logger:       logger.Named("mempool"),
	}
}

func (m *Mempool) GetPending(limit uint64) []transaction.Transaction {
	m.logger.Debug("getting transactions", "limit", limit)
	var result []transaction.Transaction

	for hash, tx := range m.transactions {
		if len(result) >= int(limit) {
			break
		}

		result = append(result, tx)
		m.logger.Info("selected transaction", "tx_id", hash.String())
	}

	return result
}

func (m *Mempool) Add(transaction transaction.Transaction) error {
	if err := m.validator.Validate(transaction); err != nil {
		return err
	}

	hash, err := transaction.Hash()

	if err != nil {
		return err
	}

	m.logger.Debug("adding transactions", "tx_id", hash.String())
	m.transactions[hash] = transaction
	m.logger.Info("added transaction", "tx_id", hash.String())

	return nil
}

func (m *Mempool) Remove(hash types.Hash) {
	m.logger.Debug("removing transactions", "tx_id", hash.String())
	delete(m.transactions, hash)
	m.logger.Info("removed transactions", "tx_id", hash.String())
}

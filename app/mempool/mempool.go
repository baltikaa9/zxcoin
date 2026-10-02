// Package mempool хранит транзакции, ожидающие включения в блок, и проверяет их валидность перед добавлением в пул.
package mempool

import (
	apptransaction "github.com/baltikaa9/zxcoin/app/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type Mempool struct {
	transactions map[types.Hash]transaction.Transaction
	validator    *apptransaction.TransactionValidator
}

func NewMempool(validator *apptransaction.TransactionValidator) *Mempool {
	return &Mempool{
		transactions: make(map[types.Hash]transaction.Transaction),
		validator:    validator,
	}
}

func (m *Mempool) GetPending(limit uint64) []transaction.Transaction {
	var result []transaction.Transaction

	for _, tx := range m.transactions {
		if len(result) >= int(limit) {
			break
		}

		result = append(result, tx)
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

	m.transactions[hash] = transaction

	return nil
}

func (m *Mempool) Remove(hash types.Hash) {
	delete(m.transactions, hash)
}

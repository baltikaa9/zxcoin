// Package mempool хранит транзакции, ожидающие включения в блок, и проверяет их валидность перед добавлением в пул.
package mempool

import (
	"zxcoin/transaction"
)

type Mempool struct {
	transactions map[[32]byte]transaction.Transaction
	tv           *transaction.TransactionValidator
}

func NewMempool(tv *transaction.TransactionValidator) *Mempool {
	return &Mempool{
		transactions: make(map[[32]byte]transaction.Transaction),
		tv:           tv,
	}
}

func (m *Mempool) GetPending(limit int) []transaction.Transaction {
	var result []transaction.Transaction

	for _, tx := range m.transactions {
		if len(result) >= limit {
			break
		}

		result = append(result, tx)
	}

	return result
}

func (m *Mempool) Add(transaction transaction.Transaction) error {
	if err := m.tv.Validate(transaction); err != nil {
		return err
	}

	hash, err := transaction.Hash()

	if err != nil {
		return err
	}

	m.transactions[hash] = transaction

	return nil
}

func (m *Mempool) Remove(hash [32]byte) {
	delete(m.transactions, hash)
}

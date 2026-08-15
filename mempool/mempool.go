package mempool

import (
	"zxcoin/transaction"
	"zxcoin/utxo"
)

type Mempool struct {
	transactions map[[32]byte]transaction.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{make(map[[32]byte]transaction.Transaction)}
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

func (m *Mempool) Add(transaction transaction.Transaction, utxoDB utxo.UTXODB) error {
	if err := transaction.Validate(utxoDB); err != nil {
		return err
	}

	m.transactions[transaction.Hash()] = transaction

	return nil
}

func (m *Mempool) Remove(hash [32]byte) {
	delete(m.transactions, hash)
}

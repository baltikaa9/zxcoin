// Package mempool хранит транзакции, ожидающие включения в блок, и проверяет их валидность перед добавлением в пул.
package mempool

import (
	"fmt"
	"sync"

	"github.com/baltikaa9/zxcoin/app/logger"
	apptransaction "github.com/baltikaa9/zxcoin/app/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type Mempool struct {
	mu           sync.RWMutex
	transactions map[types.Hash]transaction.Transaction
	inputs       map[utxo.ID]types.Hash
	validator    *apptransaction.Validator
	logger       logger.Logger
}

func NewMempool(validator *apptransaction.Validator, logger logger.Logger) *Mempool {
	return &Mempool{
		transactions: make(map[types.Hash]transaction.Transaction),
		inputs:       make(map[utxo.ID]types.Hash),
		validator:    validator,
		logger:       logger.Named("mempool"),
	}
}

func (m *Mempool) GetPending(limit uint64) []transaction.Transaction {
	m.logger.Debug("getting transactions", "limit", limit)
	var result []transaction.Transaction

	m.mu.RLock()
	defer m.mu.RUnlock()

	for hash, tx := range m.transactions {
		if uint64(len(result)) >= limit {
			break
		}

		result = append(result, tx)
		m.logger.Debug("transaction selected", "tx_id", hash.String())
	}

	return result
}

func (m *Mempool) Add(transaction transaction.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, input := range transaction.Inputs {
		if _, ok := m.inputs[input.ID]; ok {
			return DoubleSpendError{TxID: input.ID.TxID, OutIndex: input.ID.OutIndex}
		}
	}

	if err := m.validator.Validate(transaction); err != nil {
		return err
	}

	hash, err := transaction.Hash()
	if err != nil {
		return fmt.Errorf("hash transaction: %w", err)
	}

	m.logger.Debug("adding transaction", "tx_id", hash.String())
	m.transactions[hash] = transaction

	for _, input := range transaction.Inputs {
		m.inputs[input.ID] = hash
	}

	m.logger.Info("transaction added", "tx_id", hash.String())

	return nil
}

func (m *Mempool) Remove(hash types.Hash) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.remove(hash) {
		m.logger.Debug("removed transaction", "tx_id", hash.String())
	}
}

func (m *Mempool) RemoveByInput(id utxo.ID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	txHash, ok := m.inputs[id]
	if !ok {
		return
	}

	m.remove(txHash)
	m.logger.Debug("removed transaction by input", "tx_id", txHash.String(), "input_tx_id", id.TxID.String(), "out_index", id.OutIndex)
}

func (m *Mempool) remove(hash types.Hash) bool {
	tx, ok := m.transactions[hash]
	if !ok {
		return false
	}

	for _, input := range tx.Inputs {
		delete(m.inputs, input.ID)
	}

	delete(m.transactions, hash)

	return true
}

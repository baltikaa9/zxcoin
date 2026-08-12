package main

import (
	"crypto/ecdsa"
	"fmt"
	"time"
)

type Blockchain struct {
	Blocks            []Block
	CurrentDifficulty int
	CurrentAward      int
}

func (b *Blockchain) NewBlock(transactions []Transaction, creator *ecdsa.PublicKey) Block {
	coinbaseTransaction := Transaction{Outputs: []TxOutput{{b.CurrentAward, creator}}}

	prevHash := [32]byte{}

	if len(b.Blocks) > 0 {
		prevHash = b.Blocks[len(b.Blocks)-1].Header.Hash()
	}

	block := Block{
		BlockHeader{
			PrevHash:  prevHash,
			Nonce:     0,
			Timestamp: uint32(time.Now().Unix()),
		},
		append(transactions, coinbaseTransaction),
		b.CurrentDifficulty,
	}

	block.calculateRootHash()

	return block
}

func (b *Blockchain) AddBlock(block Block, utxoDB UTXODB) error {
	blockHash := block.Header.Hash()

	for i := range b.CurrentDifficulty {
		if blockHash[i] != 0 {
			return fmt.Errorf("неверный nonce")
		}
	}

	if len(b.Blocks) > 0 && block.Header.PrevHash != b.Blocks[len(b.Blocks)-1].Header.Hash() {
		return fmt.Errorf("неверный предыдущий блок")
	}

	if block.Header.RootHash != BuildMerkleTree(block.Transactions).Hash {
		return fmt.Errorf("неверный хеш корня дерева Меркла")
	}

	spentInThisBlock := make(map[UTXOKey]bool)
	foundCoinbase := false

	for _, transaction := range block.Transactions {
		if len(transaction.Inputs) == 0 && len(transaction.Outputs) == 1 && transaction.Outputs[0].Amount == b.CurrentAward {
			if foundCoinbase {
				return fmt.Errorf("больше одной coinbase-транзакции в блоке")
			}

			foundCoinbase = true
			continue
		}

		if err := transaction.Validate(utxoDB); err != nil {
			return err
		}

		for _, input := range transaction.Inputs {
			key := UTXOKey{input.TxID, input.OutIndex}

			if spentInThisBlock[key] {
				return &DoubleSpendError{input.TxID, input.OutIndex}
			}

			spentInThisBlock[key] = true
		}
	}

	for _, transaction := range block.Transactions {
		for _, input := range transaction.Inputs {
			delete(utxoDB, UTXOKey{input.TxID, input.OutIndex})
		}

		for i, output := range transaction.Outputs {
			utxoDB[UTXOKey{transaction.Hash(), i}] = UTXOEntry{output, false}
		}
	}

	b.Blocks = append(b.Blocks, block)

	return nil
}

func (b *Blockchain) MineAndAddBlock(mempool *Mempool, utxoDB UTXODB, limit int, creator *ecdsa.PublicKey) (Block, error) {
	txs := mempool.GetPending(limit)

	block := b.NewBlock(txs, creator)
	block.Mine()

	if err := b.AddBlock(block, utxoDB); err != nil {
		return Block{}, err
	}

	for _, tx := range txs {
		mempool.Remove(tx.Hash())
	}

	return block, nil
}

type DoubleSpendError struct {
	TxID     [32]byte
	OutIndex int
}

func (e *DoubleSpendError) Error() string {
	return fmt.Sprintf("UTXO (%v, %v) уже используется в данном блоке", e.TxID, e.OutIndex)
}

// Package blockchain реализует основную логику консенсуса: проверку и добавление блоков, защиту от двойной траты и управление наградой за майнинг.
package blockchain

import (
	"crypto/ecdsa"
	"time"
	"zxcoin/block"
	"zxcoin/coin"
	"zxcoin/mempool"
	"zxcoin/merkle"
	"zxcoin/transaction"
	"zxcoin/utxo"
)

type Blockchain struct {
	blocks            []block.Block
	currentDifficulty uint64
	currentAward      uint64
}

func NewBlockchain(difficulty uint64, award uint64) Blockchain {
	return Blockchain{currentDifficulty: difficulty, currentAward: award}
}

func (bc *Blockchain) newBlock(transactions []transaction.Transaction, creator *ecdsa.PublicKey) (block.Block, error) {
	coinbaseTransaction := transaction.Transaction{Outputs: []coin.TxOutput{{Amount: bc.currentAward, PublicKey: creator}}}

	prevHash := [32]byte{}

	if len(bc.blocks) > 0 {
		prevHash = bc.blocks[len(bc.blocks)-1].Header.Hash()
	}

	newBlock := block.Block{
		Header: block.BlockHeader{
			PrevHash:  prevHash,
			Nonce:     0,
			Timestamp: uint32(time.Now().Unix()),
		},
		Transactions: append(transactions, coinbaseTransaction),
		Difficulty:   bc.currentDifficulty,
	}

	err := newBlock.CalculateRootHash()

	if err != nil {
		return block.Block{}, err
	}

	return newBlock, nil
}

func (bc *Blockchain) AddBlock(block block.Block, utxoDB utxo.UTXODB) error {
	if err := bc.verifyProofOfWork(block); err != nil {
		return err
	}

	if err := bc.verifyPrevHash(block); err != nil {
		return err
	}

	if err := bc.verifyMerkleRoot(block); err != nil {
		return err
	}

	if err := bc.verifyTransactions(block, utxoDB); err != nil {
		return err
	}

	bc.applyBlock(block, utxoDB)
	bc.blocks = append(bc.blocks, block)

	return nil
}

func (bc *Blockchain) MineAndAddBlock(mempool *mempool.Mempool, utxoDB utxo.UTXODB, transactionLimit int, creator *ecdsa.PublicKey) (block.Block, error) {
	txs := mempool.GetPending(transactionLimit)

	newBlock, err := bc.newBlock(txs, creator)

	if err != nil {
		return block.Block{}, err
	}

	newBlock.Mine()

	if err := bc.AddBlock(newBlock, utxoDB); err != nil {
		return block.Block{}, err
	}

	for _, tx := range txs {
		hash, err := tx.Hash()

		if err != nil {
			return block.Block{}, err
		}

		mempool.Remove(hash)
	}

	return newBlock, nil
}

func (bc *Blockchain) verifyProofOfWork(block block.Block) error {
	blockHash := block.Header.Hash()

	for i := range bc.currentDifficulty {
		if blockHash[i] != 0 {
			return &InvalidNonceError{blockHash, bc.currentDifficulty}
		}
	}

	return nil
}

func (bc *Blockchain) verifyPrevHash(block block.Block) error {
	if len(bc.blocks) > 0 && block.Header.PrevHash != bc.blocks[len(bc.blocks)-1].Header.Hash() {
		return &InvalidPrevHashError{}
	}

	return nil
}

func (bc *Blockchain) verifyMerkleRoot(block block.Block) error {
	root, err := merkle.BuildMerkleTree(block.Transactions)

	if err != nil {
		return err
	}

	if block.Header.RootHash != root.Hash {
		return &InvalidMerkleRootError{}
	}

	return nil
}

func (bc *Blockchain) verifyTransactions(block block.Block, utxoDB utxo.UTXODB) error {
	spentInThisBlock := make(map[utxo.UTXOKey]bool)
	foundCoinbase := false

	for _, transaction := range block.Transactions {
		if len(transaction.Inputs) == 0 && len(transaction.Outputs) == 1 && transaction.Outputs[0].Amount == bc.currentAward {
			if foundCoinbase {
				return &MoreOneCoinbaseError{}
			}

			foundCoinbase = true
			continue
		}

		if err := transaction.Validate(utxoDB); err != nil {
			return err
		}

		for _, input := range transaction.Inputs {
			key := utxo.UTXOKey{TxID: input.TxID, OutIndex: input.OutIndex}

			if spentInThisBlock[key] {
				return &DoubleSpendError{input.TxID, input.OutIndex}
			}

			spentInThisBlock[key] = true
		}
	}

	return nil
}

func (bc *Blockchain) applyBlock(block block.Block, utxoDB utxo.UTXODB) error {
	for _, transaction := range block.Transactions {
		hash, err := transaction.Hash()

		if err != nil {
			return err
		}

		for _, input := range transaction.Inputs {
			delete(utxoDB, utxo.UTXOKey{TxID: input.TxID, OutIndex: input.OutIndex})
		}

		for i, output := range transaction.Outputs {
			utxoDB[utxo.UTXOKey{TxID: hash, OutIndex: uint64(i)}] = utxo.UTXOEntry{Output: output}
		}
	}

	return nil
}

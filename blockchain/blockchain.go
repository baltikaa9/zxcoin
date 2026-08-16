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
	Blocks            []block.Block
	CurrentDifficulty int
	CurrentAward      int
}

func (b *Blockchain) NewBlock(transactions []transaction.Transaction, creator *ecdsa.PublicKey) block.Block {
	coinbaseTransaction := transaction.Transaction{Outputs: []coin.TxOutput{{Amount: b.CurrentAward, PublicKey: creator}}}

	prevHash := [32]byte{}

	if len(b.Blocks) > 0 {
		prevHash = b.Blocks[len(b.Blocks)-1].Header.Hash()
	}

	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  prevHash,
			Nonce:     0,
			Timestamp: uint32(time.Now().Unix()),
		},
		Transactions: append(transactions, coinbaseTransaction),
		Difficulty:   b.CurrentDifficulty,
	}

	block.CalculateRootHash()

	return block
}

func (b *Blockchain) AddBlock(block block.Block, utxoDB utxo.UTXODB) error {
	blockHash := block.Header.Hash()

	for i := range b.CurrentDifficulty {
		if blockHash[i] != 0 {
			return &InvalidNonceError{blockHash, b.CurrentDifficulty}
		}
	}

	if len(b.Blocks) > 0 && block.Header.PrevHash != b.Blocks[len(b.Blocks)-1].Header.Hash() {
		return &InvalidPrevHashError{}
	}

	if block.Header.RootHash != merkle.BuildMerkleTree(block.Transactions).Hash {
		return &InvalidMerkleRootError{}
	}

	spentInThisBlock := make(map[utxo.UTXOKey]bool)
	foundCoinbase := false

	for _, transaction := range block.Transactions {
		if len(transaction.Inputs) == 0 && len(transaction.Outputs) == 1 && transaction.Outputs[0].Amount == b.CurrentAward {
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

	for _, transaction := range block.Transactions {
		for _, input := range transaction.Inputs {
			delete(utxoDB, utxo.UTXOKey{TxID: input.TxID, OutIndex: input.OutIndex})
		}

		for i, output := range transaction.Outputs {
			utxoDB[utxo.UTXOKey{TxID: transaction.Hash(), OutIndex: i}] = utxo.UTXOEntry{Output: output, Reserved: false}
		}
	}

	b.Blocks = append(b.Blocks, block)

	return nil
}

func (b *Blockchain) MineAndAddBlock(mempool *mempool.Mempool, utxoDB utxo.UTXODB, limit int, creator *ecdsa.PublicKey) (block.Block, error) {
	txs := mempool.GetPending(limit)

	newBlock := b.NewBlock(txs, creator)
	newBlock.Mine()

	if err := b.AddBlock(newBlock, utxoDB); err != nil {
		return block.Block{}, err
	}

	for _, tx := range txs {
		mempool.Remove(tx.Hash())
	}

	return newBlock, nil
}

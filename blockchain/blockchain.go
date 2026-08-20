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
	currentDifficulty int
	currentAward      int
}

func NewBlockchain(difficulty int, award int) Blockchain {
	return Blockchain{currentDifficulty: difficulty, currentAward: award}
}

func (bc *Blockchain) newBlock(transactions []transaction.Transaction, creator *ecdsa.PublicKey) block.Block {
	coinbaseTransaction := transaction.Transaction{Outputs: []coin.TxOutput{{Amount: bc.currentAward, PublicKey: creator}}}

	prevHash := [32]byte{}

	if len(bc.blocks) > 0 {
		prevHash = bc.blocks[len(bc.blocks)-1].Header.Hash()
	}

	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  prevHash,
			Nonce:     0,
			Timestamp: uint32(time.Now().Unix()),
		},
		Transactions: append(transactions, coinbaseTransaction),
		Difficulty:   bc.currentDifficulty,
	}

	block.CalculateRootHash()

	return block
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

	newBlock := bc.newBlock(txs, creator)
	newBlock.Mine()

	if err := bc.AddBlock(newBlock, utxoDB); err != nil {
		return block.Block{}, err
	}

	for _, tx := range txs {
		mempool.Remove(tx.Hash())
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
	if block.Header.RootHash != merkle.BuildMerkleTree(block.Transactions).Hash {
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

func (bc *Blockchain) applyBlock(block block.Block, utxoDB utxo.UTXODB) {
	for _, transaction := range block.Transactions {
		for _, input := range transaction.Inputs {
			delete(utxoDB, utxo.UTXOKey{TxID: input.TxID, OutIndex: input.OutIndex})
		}

		for i, output := range transaction.Outputs {
			utxoDB[utxo.UTXOKey{TxID: transaction.Hash(), OutIndex: i}] = utxo.UTXOEntry{Output: output}
		}
	}
}

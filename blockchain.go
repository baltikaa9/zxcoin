package main

import (
	"crypto/ecdsa"
	"fmt"
	"time"
)

type Blockchain struct {
	Blocks            []Block
	CurrentDifficulty int
}

func (b *Blockchain) NewBlock(transactions []Transaction, creator *ecdsa.PublicKey) Block {
	coinbaseTransaction := Transaction{Outputs: []TxOutput{{100, creator}}}
	
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
		append([]Transaction{coinbaseTransaction}, transactions...),
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

	for i, transaction := range block.Transactions {
		if i == 0 {
			continue
		}
		
		if err := transaction.Validate(utxoDB); err != nil {
			return err
		}

		for _, input := range transaction.Inputs {
			key := UTXOKey{input.TxID, input.OutIndex}

			if spentInThisBlock[key] {
				return fmt.Errorf("UTXO (%v, %v) уже используется в данном блоке", input.TxID, input.OutIndex)
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

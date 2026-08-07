package main

import (
	"errors"
	"fmt"
	"time"
)

type Blockchain struct {
	Blocks            []Block
	CurrentDifficulty int
}

func (b *Blockchain) NewBlock(transactions []Transaction) Block {
	prevHash := [32]byte{}
	
	if len(b.Blocks) > 0 {
		prevHash = b.Blocks[len(b.Blocks)-1].Header.Hash()
	}

	block := Block{
		BlockHeader{
			PrevHash: prevHash,
			Nonce: 0,
			Timestamp: uint32(time.Now().Unix()),
		},
		transactions,
		b.CurrentDifficulty,
	}

	block.calculateRootHash()

	return block
}

func (b *Blockchain) AddBlock(block Block, utxoDB UTXODB) error {
	if block.Header.Nonce == 0 {
		return errors.New("nonce не найден")
	}

	for _, transaction := range block.Transactions {
		inputAmount := 0
		outputAmount := 0

		for _, input := range transaction.Inputs {
			key := UTXOKey{input.TxID, input.OutIndex}
			utxo, exists := utxoDB[key]

			if !exists {
				return fmt.Errorf("UTXO (%v, %v) не найден", input.TxID, input.OutIndex)
			}

			if (input.Signature == Signature{}) || (!input.Verify(utxo.PublicKey, transaction.Hash())) {
				return fmt.Errorf("UTXO (%v, %v) не верная подпись", input.TxID, input.OutIndex)
			}

			inputAmount += utxo.Amount
		}

		for _, output := range transaction.Outputs {
			outputAmount += output.Amount
		}

		if outputAmount > inputAmount {
			return fmt.Errorf(
				"недостаточно средств: входы %d, выходы %d",
				inputAmount,
				outputAmount,
			)
		} 

		for _, input := range transaction.Inputs {
			key := UTXOKey{input.TxID, input.OutIndex}
			delete(utxoDB, key)
		}

		for i, output := range transaction.Outputs {
			utxoDB[UTXOKey{transaction.Hash(), i}] = output
		}
	}

	b.Blocks = append(b.Blocks, block)

	return nil
}

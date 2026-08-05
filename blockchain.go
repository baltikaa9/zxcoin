package main

import (
	"fmt"
)

type Blockchain struct {
	Blocks []Block
}

func (b *Blockchain) AddBlock(block Block, utxoDB UTXODB) error {
	for _, transaction := range block.Transactions {
		for _, input := range transaction.Inputs {
			key := UTXOKey{input.TxID, input.OutIndex}
			if _, exists := utxoDB[key]; !exists {
				return fmt.Errorf("UTXO (%v, %v) не найден", input.TxID, input.OutIndex)
			}

			delete(utxoDB, key)
		}

		for i, output := range transaction.Outputs {
			utxoDB[UTXOKey{transaction.Hash(), i}] = output
		}
	}

	b.Blocks = append(b.Blocks, block)

	return nil
}

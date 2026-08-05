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
			if !ValidateTransactionInput(input, utxoDB) {
				return fmt.Errorf("UTXO (%v, %v) не найден", input.TxID, input.OutIndex)
			}
		}
	}

	b.Blocks = append(b.Blocks, block)

	UpdateUtxoDB(block.Transactions, &utxoDB)

	return nil
}

func ValidateTransactionInput(input TxInput, utxoDB UTXODB) bool {
	key := UTXOKey{input.TxID, input.OutIndex}
	_, exists := utxoDB[key]

	return exists
}

func UpdateUtxoDB(transactions []Transaction, utxoDB *UTXODB) {
	for _, transaction := range transactions {
		for _, input := range transaction.Inputs {
			delete(*utxoDB, UTXOKey{input.TxID, input.OutIndex})
		}

		for i, output := range transaction.Outputs {
			(*utxoDB)[UTXOKey{transaction.Hash(), i}] = output
		}
	}
}

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type TxInput struct {
	TxID      string
	OutIndex  int
	Signature string
}

type TxOutput struct {
	Amount    int
	PublicKey string
}

type Transaction struct {
	Inputs  []TxInput
	Outputs []TxOutput
}

func NewTransaction(inputs []TxInput, outputs []TxOutput) Transaction {
	t := Transaction{inputs, outputs}
	return t
}

func (t Transaction) Hash() string {
	data := fmt.Sprintf("%v", t.Inputs) + fmt.Sprintf("%v", t.Outputs)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

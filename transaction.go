package main

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
	Hash string
	Inputs []TxInput
	Outputs []TxOutput
}
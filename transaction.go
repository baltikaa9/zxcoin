package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

type TxInput struct {
	TxID      []byte
	OutIndex  int
	PublicKey *ecdsa.PublicKey
	Signature []byte
}

type TxOutput struct {
	Amount    int
	PublicKey *ecdsa.PublicKey
}

type Transaction struct {
	Inputs  []TxInput
	Outputs []TxOutput
}

func NewTransaction(inputs []TxInput, outputs []TxOutput) Transaction {
	t := Transaction{inputs, outputs}
	return t
}

func (t Transaction) Hash() []byte {
	for i := range t.Inputs {
		t.Inputs[i].Signature = nil
	}

	data := fmt.Sprintf("%v", t)
	hash := sha256.Sum256([]byte(data))

	return hash[:]
}

func (in *TxInput) Sign(privateKey *ecdsa.PrivateKey, hash []byte) error {
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return err
	}

	sign := append(r.Bytes(), s.Bytes()...)
	in.Signature = sign

	return nil
}

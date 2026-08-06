package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

type TxInput struct {
	TxID      [32]byte
	OutIndex  int
	Signature Signature
}

type TxOutput struct {
	Amount int
	PublicKey *ecdsa.PublicKey
}

type Transaction struct {
	Inputs  []TxInput
	Outputs []TxOutput
}

func NewTransaction(inputs []TxInput, outputs []TxOutput) Transaction {
	return Transaction{inputs, outputs}
}

func (t Transaction) Hash() [32]byte {
	tmp := t

	tmp.Inputs = make([]TxInput, len(t.Inputs))
	copy(tmp.Inputs, t.Inputs)

	for i := range tmp.Inputs {
		tmp.Inputs[i].Signature = Signature{}
	}

	data := fmt.Sprintf("%v", tmp)
	hash := sha256.Sum256([]byte(data))

	return hash
}

func (in *TxInput) Sign(privateKey *ecdsa.PrivateKey, hash [32]byte) error {
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return err
	}

	in.Signature = Signature{r, s}

	return nil
}

func (in *TxInput) Verify(publicKey *ecdsa.PublicKey, hash [32]byte) bool {
	return ecdsa.Verify(publicKey, hash[:], in.Signature.R, in.Signature.S)
}

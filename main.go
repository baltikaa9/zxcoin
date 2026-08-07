package main

import (
	// "crypto/ecdsa"
	"fmt"
	// "math/big"
	// "time"
)

func main() {
	myWallet := NewWallet()
	otherWallet := NewWallet()

	fmt.Printf("myWallet: %v, otherWallet: %v\n", myWallet, otherWallet)

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: TxOutput{5, myWallet.PublicKey},
	}

	t := NewTransaction(
		[]TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		[]TxOutput{{100, otherWallet.PublicKey}},
	)

	tHash := t.Hash()

	for i := range t.Inputs {
		t.Inputs[i].Sign(myWallet.PrivateKey, tHash)
		// t.Inputs[i].Signature = Signature{big.NewInt(1), big.NewInt(2)}
	}

	block := NewBlock(Block{}, []Transaction{t})
	block.Header.Nonce = 10

	b := Blockchain{}
	
	fmt.Printf("%v\n", utxoDB)

	err := b.AddBlock(block, utxoDB)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", utxoDB)

	// fmt.Printf("%v\n", t.Inputs[0].Signature)

	// oldOutput := t0.Outputs[t.Inputs[0].OutIndex]

	// fmt.Printf("%v", t.Inputs[0].Verify(oldOutput.PublicKey, tHash))
}

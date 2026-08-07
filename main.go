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

	fmt.Printf("myWallet: %v, otherWallet: %v\n\n", myWallet, otherWallet)

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: TxOutput{5, myWallet.PublicKey},
		UTXOKey{[32]byte{}, 1}: TxOutput{3, myWallet.PublicKey},
		UTXOKey{[32]byte{}, 2}: TxOutput{1, myWallet.PublicKey},
		UTXOKey{[32]byte{}, 3}: TxOutput{10, myWallet.PublicKey},
	}

	t, err := myWallet.CreateTransaction(otherWallet.PublicKey, 10, utxoDB)
	t1, err1 := myWallet.CreateTransaction(otherWallet.PublicKey, 10, utxoDB)

	if err != nil {
		panic(err)
	}

	if err1 != nil {
		panic(err1)
	}

	block := NewBlock(Block{}, []Transaction{t})
	block.Header.Nonce = 10

	block1 := NewBlock(block, []Transaction{t1})
	block1.Header.Nonce = 10

	b := Blockchain{}

	fmt.Printf("%v\n\n", utxoDB)

	err = b.AddBlock(block, utxoDB)
	err1 = b.AddBlock(block1, utxoDB)

	if err != nil {
		panic(err)
	}

	if err1 != nil {
		panic(err1)
	}

	fmt.Printf("%v\n", utxoDB)
}

package main

import (
	"fmt"
)

func main() {
	myWallet := NewWallet()
	otherWallet := NewWallet()

	fmt.Printf("myWallet: %v, otherWallet: %v\n\n", myWallet, otherWallet)

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: TxOutput{5, myWallet.PublicKey},
		UTXOKey{[32]byte{}, 1}: TxOutput{3, myWallet.PublicKey},
		UTXOKey{[32]byte{}, 2}: TxOutput{11, myWallet.PublicKey},
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
	
	b := Blockchain{CurrentDifficulty: 2}

	block := b.NewBlock([]Transaction{t, t1})
	// block.Header.Nonce = 10
	block.Mine()

	// fmt.Printf("nonce = %v\n", block.Header.Nonce)
	// fmt.Printf("root = %v\n", block.Header.RootHash)
	// fmt.Printf("prev = %v\n", block.Header.PrevHash)
	err = b.AddBlock(block, utxoDB)
	
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n%v\n\n", utxoDB)
	
	

	// block1 := b.NewBlock([]Transaction{t1})
	// block1.Header.Nonce = 10
	// block1.Mine()
	
	// fmt.Printf("nonce = %v\n", block1.Header.Nonce)
	// fmt.Printf("root = %v\n", block1.Header.RootHash)
	// fmt.Printf("prev = %v\n", block1.Header.PrevHash)

	// fmt.Printf("%v\n\n", utxoDB)

	// err1 = b.AddBlock(block1, utxoDB)

	// if err1 != nil {
		// panic(err1)
	// }

	fmt.Printf("%v\n", utxoDB)
}

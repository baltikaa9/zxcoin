package main

import (
	"fmt"
)

func main() {
	myWallet := NewWallet()
	otherWallet := NewWallet()

	fmt.Printf("myWallet: %v, otherWallet: %v\n\n", myWallet, otherWallet)

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: UTXOEntry{TxOutput{5, myWallet.PublicKey}, false},
		UTXOKey{[32]byte{}, 1}: UTXOEntry{TxOutput{3, myWallet.PublicKey}, false},
		UTXOKey{[32]byte{}, 2}: UTXOEntry{TxOutput{11, myWallet.PublicKey}, false},
		UTXOKey{[32]byte{}, 3}: UTXOEntry{TxOutput{10, myWallet.PublicKey}, false},
	}

	// var txPool []Transaction

	t, err := myWallet.CreateTransaction(otherWallet.PublicKey, 10, utxoDB)
	t1, err1 := myWallet.CreateTransaction(otherWallet.PublicKey, 10, utxoDB)
	// t2, _ := myWallet.CreateTransaction(otherWallet.PublicKey, 1, utxoDB)
	// t3, _ := myWallet.CreateTransaction(otherWallet.PublicKey, 1, utxoDB)
	// t4, _ := myWallet.CreateTransaction(otherWallet.PublicKey, 1, utxoDB)

	if err != nil {
		panic(err)
	}

	if err1 != nil {
		panic(err1)
	}
	fmt.Printf("было\n%v\n\n", utxoDB)

	mempool := NewMempool()

	err = mempool.Add(t, utxoDB)
	if err != nil {
		panic(err)
	}
	
	err = mempool.Add(t1, utxoDB)
	if err != nil {
		panic(err)
	}

	b := Blockchain{CurrentDifficulty: 2, CurrentAward: 100}

	b.MineAndAddBlock(mempool, utxoDB, 3, myWallet.PublicKey)

	// txs := mempool.GetPending(3)

	// block := b.NewBlock(txs, myWallet.PublicKey)
	// block.Header.Nonce = 10
	// block.Mine()

	// fmt.Printf("nonce = %v\n", block.Header.Nonce)
	// fmt.Printf("root = %v\n", block.Header.RootHash)
	// fmt.Printf("prev = %v\n", block.Header.PrevHash)
	// err = b.AddBlock(block, utxoDB)

	// if err != nil {
		// panic(err)
	// }

	// for _, tx := range txs {
		// mempool.Remove(tx.Hash())
	// }

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

	fmt.Printf("стало\n%v\n", utxoDB)
}

package main

import (
	"fmt"
	"zxcoin/blockchain"
	"zxcoin/coin"
	"zxcoin/mempool"
	"zxcoin/utxo"
	"zxcoin/wallet"
)

func main() {
	myWallet := wallet.NewWallet()
	otherWallet := wallet.NewWallet()

	fmt.Printf("myWallet: %v, otherWallet: %v\n\n", myWallet, otherWallet)

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{TxID: [32]byte{}, OutIndex: 0}: utxo.UTXOEntry{Output: coin.TxOutput{Amount: 5, PublicKey: myWallet.PublicKey}},
		// utxo.UTXOKey{TxID: [32]byte{}, OutIndex: 1}: utxo.UTXOEntry{Output: coin.TxOutput{Amount: 3, PublicKey: myWallet.PublicKey}},
		// utxo.UTXOKey{TxID: [32]byte{}, OutIndex: 2}: utxo.UTXOEntry{Output: coin.TxOutput{Amount: 11, PublicKey: myWallet.PublicKey}},
		// utxo.UTXOKey{TxID: [32]byte{}, OutIndex: 3}: utxo.UTXOEntry{Output: coin.TxOutput{Amount: 10, PublicKey: myWallet.PublicKey}},
	}

	// t1, err := myWallet.CreateTransaction(otherWallet.PublicKey, 10, utxoDB)
	// if err != nil {
	// panic(err)
	// }

	t2, err := myWallet.CreateTransaction(otherWallet.PublicKey, 0, utxoDB)
	if err != nil {
		panic(err)
	}

	fmt.Printf("было\n%v\n\n", utxoDB)

	mp := mempool.NewMempool()

	// if err := mp.Add(t1, utxoDB); err != nil {
	// panic(err)
	// }
	if err := mp.Add(t2, utxoDB); err != nil {
		panic(err)
	}

	bc := blockchain.Blockchain{CurrentDifficulty: 2, CurrentAward: 42}

	if _, err := bc.MineAndAddBlock(mp, utxoDB, 3, myWallet.PublicKey); err != nil {
		panic(err)
	}

	fmt.Printf("стало\n%v\n", utxoDB)
}

package main

import (
	"fmt"
	"zxcoin/blockchain"
	"zxcoin/mempool"
	"zxcoin/utxo"
	"zxcoin/utxo/inmemory"
	"zxcoin/wallet"
)

func main() {
	repo := inmemory.NewRepository()
	tracker := wallet.NewReservationTracker()

	myWallet := wallet.NewWallet(tracker, repo)
	otherWallet := wallet.NewWallet(tracker, repo)

	fmt.Printf("myWallet: %v, otherWallet: %v\n\n", myWallet, otherWallet)

	utxos := []utxo.UTXO{
		{ID: utxo.UTXOID{TxID: [32]byte{}, OutIndex: 0}, Amount: 5, Owner: myWallet.PublicKey},
		{ID: utxo.UTXOID{TxID: [32]byte{}, OutIndex: 1}, Amount: 3, Owner: myWallet.PublicKey},
		{ID: utxo.UTXOID{TxID: [32]byte{}, OutIndex: 2}, Amount: 11, Owner: myWallet.PublicKey},
		{ID: utxo.UTXOID{TxID: [32]byte{}, OutIndex: 3}, Amount: 10, Owner: myWallet.PublicKey},
	}
	for _, u := range utxos {
		err := repo.Save(u)

		if err != nil {
			panic(err)
		}
	}
	mp := mempool.NewMempool()

	t1, err := myWallet.CreateTransaction(otherWallet.PublicKey, 10)

	if err != nil {
		panic(err)
	}

	t2, err := myWallet.CreateTransaction(otherWallet.PublicKey, 1)

	if err != nil {
		panic(err)
	}

	my, _ := repo.FindByOwner(myWallet.PublicKey)
	other, _ := repo.FindByOwner(otherWallet.PublicKey)
	fmt.Printf("было\nMy: %v\nOther: %v\n\n", my, other)

	if err := mp.Add(t1, repo); err != nil {
		panic(err)
	}

	if err := mp.Add(t2, repo); err != nil {
		panic(err)
	}

	bc := blockchain.NewBlockchain(2, 42)

	if _, err := bc.MineAndAddBlock(mp, repo, 3, myWallet.PublicKey); err != nil {
		panic(err)
	}

	my, _ = repo.FindByOwner(myWallet.PublicKey)
	other, _ = repo.FindByOwner(otherWallet.PublicKey)
	fmt.Printf("стало\nMy: %v\nOther: %v\n\n", my, other)
}

package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/baltikaa9/zxcoin/adapters/logger"
	"github.com/baltikaa9/zxcoin/adapters/persistence/inmemory"
	"github.com/baltikaa9/zxcoin/adapters/persistence/sqlite"
	appblockchain "github.com/baltikaa9/zxcoin/app/blockchain"
	"github.com/baltikaa9/zxcoin/app/mempool"
	"github.com/baltikaa9/zxcoin/app/transaction"
	"github.com/baltikaa9/zxcoin/app/wallet"
	"github.com/baltikaa9/zxcoin/core/domain/blockchain"

	_ "modernc.org/sqlite"
)

func main() {
	logger, err := logger.NewZapLogger()

	if err != nil {
		log.Fatal(err)
	}

	defer logger.Sync()

	db, err := sql.Open("sqlite", "zxcoin.db")

	if err != nil {
		panic(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	_ = inmemory.NewRepository()
	repo := sqlite.NewRepository(db)

	tracker := wallet.NewReservationTracker()
	validator := transaction.NewValidator(repo)
	mp := mempool.NewMempool(validator, logger)
	bc := blockchain.NewBlockchain(2, 42)

	bs := appblockchain.NewBlockchainService(bc, repo, validator, mp, logger)

	myWalletKey, err := wallet.LoadPrivateKey("wallet.key")

	if err != nil {
		panic(err)
	}

	otherWalletKey, err := wallet.LoadPrivateKey("other.key")

	if err != nil {
		panic(err)
	}

	myWallet := wallet.LoadWallet(myWalletKey, tracker, repo, logger)
	otherWallet := wallet.LoadWallet(otherWalletKey, tracker, repo, logger)

	pk, _ := myWallet.PublicKey.Bytes()

	fmt.Println("key:", hex.EncodeToString(pk))

	// fmt.Printf("myWallet: %v, otherWallet: %v\n\n", myWallet, otherWallet)

	// coins := []coin.TxOutput{
	// {Amount: 5, Owner: myWallet.PublicKey},
	// {Amount: 3, Owner: myWallet.PublicKey},
	// {Amount: 11, Owner: myWallet.PublicKey},
	// {Amount: 10, Owner: myWallet.PublicKey},
	// }

	// utxos := make([]utxo.UTXO, 0, len(coins))

	// 	for i, c := range coins {
	// 		utxos = append(utxos, utxo.UTXO{
	// 			ID:     utxo.UTXOID{TxID: types.Hash{}, OutIndex: uint64(i)},
	// 			Output: c,
	// 		})
	// 	}
	//
	// 	for _, u := range utxos {
	// 		err := repo.Save(u)
	//
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 	}

	t1, err := myWallet.CreateTransaction(otherWallet.PublicKey, 10)

	if err != nil {
		panic(err)
	}

	for _, i := range t1.Inputs {
		fmt.Printf("txID: %v\noutIndex: %v\nsignature: %v\n\n", i.ID.TxID, i.ID.OutIndex, hex.EncodeToString(i.Signature.Marshal()))
	}

	for _, o := range t1.Outputs {
		ownerBytes, err := o.Owner.Bytes()
		if err != nil {
			panic(err)
		}
		fmt.Printf("amount: %v\nowner: %v\n\n", o.Amount, hex.EncodeToString(ownerBytes))
	}

	// for _, i := range t1.Inputs {
	// fmt.Println(t1.Hash())
	// fmt.Println(string(i.Signature.R.Bytes()))
	// fmt.Println(string(i.Signature.S.Bytes()))
	// }

	// t2, err := myWallet.CreateTransaction(otherWallet.PublicKey, 0)

	// if err != nil {
	// panic(err)
	// }

	my, _ := repo.FindByOwner(myWallet.PublicKey)
	other, _ := repo.FindByOwner(otherWallet.PublicKey)
	fmt.Printf("было\nMy: %v\nOther: %v\n\n", my, other)

	if err := mp.Add(t1); err != nil {
		panic(err)
	}

	// if err := mp.Add(t2); err != nil {
	// panic(err)
	// }

	if _, err := bs.MineAndAddBlock(0, myWallet.PublicKey); err != nil {
		panic(err)
	}

	my, _ = repo.FindByOwner(myWallet.PublicKey)
	other, _ = repo.FindByOwner(otherWallet.PublicKey)
	fmt.Printf("стало\nMy: %v\nOther: %v\n\n", my, other)
}

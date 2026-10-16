package main

import (
	"context"
	"fmt"
	"log"
	"os"

	adapterlogger "github.com/baltikaa9/zxcoin/adapters/logger"
	"github.com/baltikaa9/zxcoin/adapters/persistence/inmemory"
	appblockchain "github.com/baltikaa9/zxcoin/app/blockchain"
	applogger "github.com/baltikaa9/zxcoin/app/logger"
	"github.com/baltikaa9/zxcoin/app/mempool"
	"github.com/baltikaa9/zxcoin/app/transaction"
	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/app/wallet"
	"github.com/baltikaa9/zxcoin/core/domain/blockchain"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

func main() {
	lg, err := adapterlogger.NewZapLogger("logs/zxcoin.json")

	if err != nil {
		log.Fatal(err)
	}

	err = run(lg)

	if err != nil {
		lg.Error("demo failed", "err", err)
	}

	_ = lg.Sync()

	if err != nil {
		os.Exit(1)
	}
}

func run(lg applogger.Logger) error {
	lg.Info("started")
	repo := inmemory.NewRepository()
	tracker := wallet.NewReservationTracker()
	validator := transaction.NewValidator(repo)
	mp := mempool.NewMempool(validator, lg)
	bc := blockchain.NewBlockchain(4, 42)
	bs := appblockchain.NewService(bc, repo, validator, mp, lg)

	myWallet := wallet.NewWallet(tracker, repo, lg)
	otherWallet := wallet.NewWallet(tracker, repo, lg)

	for i, amount := range []uint64{5, 3, 11, 10} {
		u := utxo.UTXO{
			ID:     utxo.ID{TxID: types.Hash{}, OutIndex: uint64(i)},
			Output: coin.TxOutput{Amount: amount, Owner: myWallet.PublicKey},
		}

		if err := repo.Save(u); err != nil {
			return fmt.Errorf("save utxo: %w", err)
		}
	}

	fmt.Println("было")

	if err := printBalances(repo, myWallet, otherWallet); err != nil {
		return err
	}

	t1, err := myWallet.CreateTransaction(otherWallet.PublicKey, 10)

	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	if err := mp.Add(t1); err != nil {
		return fmt.Errorf("add to mempool: %w", err)
	}

	if _, err := bs.MineAndAddBlock(context.Background(), 10, myWallet.PublicKey); err != nil {
		return fmt.Errorf("mine block: %w", err)
	}

	fmt.Println("стало")

	return printBalances(repo, myWallet, otherWallet)
}

func printBalances(repo apputxo.Repository, my, other wallet.Wallet) error {
	for name, w := range map[string]wallet.Wallet{"my": my, "other": other} {
		utxos, err := repo.FindByOwner(w.PublicKey)

		if err != nil {
			return fmt.Errorf("find utxo of %s: %w", name, err)
		}

		var sum uint64

		for _, u := range utxos {
			sum += u.Output.Amount
		}

		fmt.Printf("  %s (%s): %d UTXO, баланс %d\n", name, w.PublicKey.Short(), len(utxos), sum)
	}

	return nil
}

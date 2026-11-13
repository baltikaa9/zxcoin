package main

import (
	"context"
	"fmt"
	"log"
	"os"

	adapterlogger "github.com/baltikaa9/zxcoin/adapters/logger"
	"github.com/baltikaa9/zxcoin/adapters/persistence/sqlite"
	appblockchain "github.com/baltikaa9/zxcoin/app/blockchain"
	applogger "github.com/baltikaa9/zxcoin/app/logger"
	"github.com/baltikaa9/zxcoin/app/mempool"
	apptransaction "github.com/baltikaa9/zxcoin/app/transaction"
	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/app/wallet"
	"github.com/baltikaa9/zxcoin/core/domain/blockchain"
	"github.com/baltikaa9/zxcoin/core/testutil"
	_ "modernc.org/sqlite"
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

	db, err := testutil.OpenTestSQLiteDB("migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	defer db.Close()

	utxoRepo := sqlite.NewUTXO(db)
	blockRepo := sqlite.NewBlock(db)
	tracker := wallet.NewReservationTracker()
	validator := apptransaction.NewValidator(utxoRepo)
	mp := mempool.NewMempool(validator, lg)
	bc := blockchain.NewBlockchain(2, 42)
	bs := appblockchain.NewService(bc, blockRepo, validator, mp, lg)

	myKey, err := wallet.LoadPrivateKey("wallet.key.test")
	if err != nil {
		return fmt.Errorf("load key: %w", err)
	}
	otherKey, err := wallet.LoadPrivateKey("other.key.test")
	if err != nil {
		return fmt.Errorf("load key: %w", err)
	}
	myWallet := wallet.LoadWallet(myKey, tracker, utxoRepo, lg)
	otherWallet := wallet.LoadWallet(otherKey, tracker, utxoRepo, lg)

	fmt.Println("было")

	if err := printBalances(utxoRepo, myWallet, otherWallet); err != nil {
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

	return printBalances(utxoRepo, myWallet, otherWallet)
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

package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/baltikaa9/zxcoin/adapters/http/handlers"
	"github.com/baltikaa9/zxcoin/adapters/http/middleware"
	"github.com/baltikaa9/zxcoin/adapters/logger"
	"github.com/baltikaa9/zxcoin/adapters/persistence/sqlite"
	appblockchain "github.com/baltikaa9/zxcoin/app/blockchain"
	"github.com/baltikaa9/zxcoin/app/mempool"
	"github.com/baltikaa9/zxcoin/app/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/blockchain"
	_ "modernc.org/sqlite"
)

func main() {
	lg, err := logger.NewZapLogger("logs/http.json")

	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = lg.Sync() }()

	db, err := sql.Open("sqlite", "file:zxcoin.db?_pragma=foreign_keys(1)")

	if err != nil {
		panic(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	repo := sqlite.NewUTXO(db)
	validator := transaction.NewValidator(repo)
	mp := mempool.NewMempool(validator, lg)
	bc := blockchain.NewBlockchain(2, 42)
	bs := appblockchain.NewService(bc, repo, validator, mp, lg)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /transactions", handlers.NewMempool(mp, lg).AddTransaction)
	mux.HandleFunc("GET /utxo/{owner}", handlers.NewUTXO(repo, lg).GetUTXOByOwner)
	mux.HandleFunc("POST /mine", handlers.NewMining(bs, lg).Mine)

	handler := middleware.Logging(lg.Named("http"))(
		middleware.Recovery(lg.Named("http"))(
			middleware.JSON(mux),
		),
	)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}

package main

import (
	"database/sql"
	"net/http"

	"github.com/baltikaa9/zxcoin/adapters/http/handlers"
	"github.com/baltikaa9/zxcoin/adapters/http/middleware"
	"github.com/baltikaa9/zxcoin/adapters/persistence/sqlite"
	"github.com/baltikaa9/zxcoin/app/mempool"
	"github.com/baltikaa9/zxcoin/app/transaction"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "zxcoin.db")

	if err != nil {
		panic(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	repo := sqlite.NewRepository(db)
	validator := transaction.NewValidator(repo)
	mp := mempool.NewMempool(validator)

	mux := http.NewServeMux()

	// mux.HandleFunc("POST /health", nil)
	mux.HandleFunc("POST /transactions", handlers.NewMempoolHandler(mp).AddTransaction)
	// mux.HandleFunc("POST /mine", nil)

	if err := http.ListenAndServe(":8080", middleware.JSONMiddleware(mux)); err != nil {
		panic(err)
	}
}

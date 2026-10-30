package main

import (
	"log"

	"github.com/baltikaa9/zxcoin/adapters/logger"
	"github.com/baltikaa9/zxcoin/adapters/persistence/inmemory"
	"github.com/baltikaa9/zxcoin/app/wallet"
)

func main() {
	lg, err := logger.NewZapLogger("logs/wallet.json")

	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = lg.Sync() }()

	w := wallet.NewWallet(wallet.NewReservationTracker(), inmemory.NewBlock(), lg)

	if err := wallet.SavePrivateKey("other.key", w.PrivateKey); err != nil {
		log.Fatal(err)
	}
}

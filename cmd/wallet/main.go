package main

import (
	"log"

	"github.com/baltikaa9/zxcoin/adapters/logger"
	"github.com/baltikaa9/zxcoin/adapters/persistence/inmemory"
	"github.com/baltikaa9/zxcoin/app/wallet"
)

func main() {
	lg, err := logger.NewZapLogger()

	if err != nil {
		log.Fatal(err)
	}

	defer lg.Sync()

	w := wallet.NewWallet(wallet.NewReservationTracker(), inmemory.NewRepository(), lg)

	wallet.SavePrivateKey("other.key", w.PrivateKey)
}

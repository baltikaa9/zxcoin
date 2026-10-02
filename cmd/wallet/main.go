package main

import (
	"github.com/baltikaa9/zxcoin/adapters/persistence/inmemory"
	"github.com/baltikaa9/zxcoin/app/wallet"
)

func main() {
	w := wallet.NewWallet(wallet.NewReservationTracker(), inmemory.NewRepository())

	wallet.SavePrivateKey("other.key", w.PrivateKey)
}

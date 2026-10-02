package main

import (
	"github.com/baltikaa9/zxcoin/utxo/inmemory"
	"github.com/baltikaa9/zxcoin/wallet"
)

func main() {
	w := wallet.NewWallet(wallet.NewReservationTracker(), inmemory.NewRepository())

	wallet.SavePrivateKey("other.key", w.PrivateKey)
}

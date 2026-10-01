package main

import (
	"zxcoin/utxo/inmemory"
	"zxcoin/wallet"
)

func main() {
	w := wallet.NewWallet(wallet.NewReservationTracker(), inmemory.NewRepository())

	wallet.SavePrivateKey("other.key", w.PrivateKey)
}

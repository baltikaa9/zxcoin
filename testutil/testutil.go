package testutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"zxcoin/coin"
	"zxcoin/utxo"
)

func GenerateKeyPair(t *testing.T) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		t.Fatalf("не удалось сгенерировать ключ: %v", err)
	}

	return privateKey, &privateKey.PublicKey
}

func GenerateSingleUtxo(t *testing.T, amount int, publicKey *ecdsa.PublicKey) utxo.UTXODB {
	return utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    amount,
				PublicKey: publicKey,
			},
		},
	}
}

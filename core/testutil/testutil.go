// Package testutil предоставляет вспомогательные функции для тестов: генерацию ключевых пар и тестовых наборов UTXO.
package testutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

func GenerateKeyPair(t *testing.T) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		t.Fatalf("не удалось сгенерировать ключ: %v", err)
	}

	return privateKey, &privateKey.PublicKey
}

func GenerateSingleUtxo(t *testing.T, amount uint64, publicKey *ecdsa.PublicKey, repo apputxo.Repository) {
	t.Helper()

	u := utxo.UTXO{
		ID: utxo.UTXOID{},
		Output: coin.TxOutput{
			Amount: amount,
			Owner:  publicKey,
		},
	}

	if err := repo.Save(u); err != nil {
		t.Fatalf("не удалось сохранить UTXO: %v", err)
	}
}

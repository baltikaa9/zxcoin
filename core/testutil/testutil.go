// Package testutil предоставляет вспомогательные функции для тестов: генерацию ключевых пар и тестовых наборов UTXO.
package testutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type utxoSaver interface {
	Save(u utxo.UTXO) error
}

func GenerateKeyPair(t *testing.T) (*ecdsa.PrivateKey, types.PublicKey) {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		t.Fatalf("не удалось сгенерировать ключ: %v", err)
	}

	return privateKey, types.PublicKey{PublicKey: &privateKey.PublicKey}
}

func GenerateSingleUtxo(t *testing.T, amount uint64, publicKey types.PublicKey, repo utxoSaver) {
	t.Helper()

	u := utxo.UTXO{
		ID: utxo.ID{},
		Output: coin.TxOutput{
			Amount: amount,
			Owner:  publicKey,
		},
	}

	if err := repo.Save(u); err != nil {
		t.Fatalf("не удалось сохранить UTXO: %v", err)
	}
}

func CheckError[E error](t *testing.T, err error) {
	t.Helper()

	if _, ok := errors.AsType[E](err); ok {
		return
	}

	actual := err

	for errors.Unwrap(actual) != nil {
		actual = errors.Unwrap(actual)
	}

	t.Fatalf("ожидался тип ошибки %T, получен %T: %v", *new(E), actual, err)
}

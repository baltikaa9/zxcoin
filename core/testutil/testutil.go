// Package testutil предоставляет вспомогательные функции для тестов: генерацию ключевых пар и тестовых наборов UTXO.
package testutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type utxoSaver interface {
	Save(u utxo.UTXO)
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

	repo.Save(utxo.UTXO{
		ID: utxo.ID{},
		Output: coin.TxOutput{
			Amount: amount,
			Owner:  publicKey,
		},
	})
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

func OpenTestSQLiteDB(migrationsPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file::memory:?_pragma=foreign_keys(1)")

	if err != nil {
		return nil, err
	}

	// Каждое соединение с :memory: получает свою пустую базу.
	db.SetMaxOpenConns(1)

	migrations, err := filepath.Glob(migrationsPath)

	if err != nil {
		db.Close()
		return nil, err
	}

	for _, path := range migrations {
		migration, err := os.ReadFile(path)

		if err != nil {
			db.Close()
			return nil, err
		}

		if _, err := db.Exec(string(migration)); err != nil {
			db.Close()
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}

	return db, nil
}

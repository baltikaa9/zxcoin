package coin_test

import (
	"errors"
	"math/big"
	"testing"
	"zxcoin/coin"
	"zxcoin/testutil"
)

func TestSerialize_NilPublicKey(t *testing.T) {
	output := coin.TxOutput{Amount: 1}

	_, err := output.Serialize()

	if _, ok := errors.AsType[coin.NilPublicKeyError](err); !ok {
		t.Fatalf("ожидалась NilPublicKeyError, получено: %v", err)
	}
}

func TestSerialize_PublicKeySerialize(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	publicKey.X = big.NewInt(42)
	output := coin.TxOutput{Amount: 1, PublicKey: publicKey}

	_, err := output.Serialize()

	if _, ok := errors.AsType[coin.PublicKeySerializeError](err); !ok {
		t.Fatalf("ожидалась PublicKeySerializeError, получено: %v", err)
	}
}

func TestSerialize_Success(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	output := coin.TxOutput{Amount: 1, PublicKey: publicKey}

	_, err := output.Serialize()

	if err != nil {
		t.Fatal(err)
	}
}

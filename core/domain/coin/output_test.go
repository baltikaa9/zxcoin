package coin_test

import (
	"math/big"
	"testing"

	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/testutil"
)

func TestSerialize_NilPublicKey(t *testing.T) {
	output := coin.TxOutput{Amount: 1}

	_, err := output.Serialize()

	testutil.CheckError[coin.NilPublicKeyError](t, err)
}

func TestSerialize_PublicKeySerialize(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	publicKey.X = big.NewInt(42)
	output := coin.TxOutput{Amount: 1, Owner: publicKey}

	_, err := output.Serialize()

	testutil.CheckError[coin.PublicKeySerializeError](t, err)
}

func TestSerialize_Success(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	output := coin.TxOutput{Amount: 1, Owner: publicKey}

	_, err := output.Serialize()

	if err != nil {
		t.Fatal(err)
	}
}

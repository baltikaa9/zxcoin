package types

import (
	"crypto/ecdsa"
	"encoding/hex"
)

type PublicKey struct {
	*ecdsa.PublicKey
}

func (k PublicKey) Short() string {
	if k.PublicKey == nil {
		return "nil"
	}

	b, err := k.Bytes()

	if err != nil || len(b) < 5 {
		return "invalid"
	}

	return hex.EncodeToString(b[1:5])
}

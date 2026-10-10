package types

import (
	"crypto/ecdsa"
	"encoding/hex"
)

type PublicKey struct {
	*ecdsa.PublicKey
}

func (k PublicKey) Short() string {
	b, _ := k.Bytes()
	return hex.EncodeToString(b[1:5])
}

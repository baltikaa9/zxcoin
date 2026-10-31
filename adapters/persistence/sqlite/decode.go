package sqlite

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/domain/types"
)

func decodeHash(b []byte) (types.Hash, error) {
	if len(b) != len(types.Hash{}) {
		return types.Hash{}, fmt.Errorf(
			"invalid hash length: got %d, want %d",
			len(b),
			len(types.Hash{}),
		)
	}

	return types.Hash(b), nil
}

func decodePublicKey(b []byte) (types.PublicKey, error) {
	key, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), b)

	if err != nil {
		return types.PublicKey{}, err
	}

	return types.PublicKey{PublicKey: key}, nil
}

func decodeHeader(prevHash, rootHash []byte, h *block.Header) error {
	var err error

	if h.PrevHash, err = decodeHash(prevHash); err != nil {
		return err
	}

	if h.RootHash, err = decodeHash(rootHash); err != nil {
		return err
	}

	return err
}

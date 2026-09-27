// Package coin определяет базовый примитив ценности в системе — выход транзакции (TxOutput).
package coin

import (
	"crypto/ecdsa"
	"encoding/binary"
)

type TxOutput struct {
	Amount    uint64
	PublicKey *ecdsa.PublicKey
}

func (o TxOutput) Serialize() ([]byte, error) {
	if o.PublicKey == nil {
		return nil, NilPublicKeyError{}
	}

	key, err := o.PublicKey.Bytes()

	if err != nil {
		return nil, PublicKeySerializeError{Key: o.PublicKey, Previous: err}
	}

	buf := binary.BigEndian.AppendUint32(nil, uint32(len(key)))
	buf = append(buf, key...)
	buf = binary.BigEndian.AppendUint64(buf, o.Amount)

	return buf, nil
}

// Package coin определяет базовый примитив ценности в системе — выход транзакции (TxOutput).
package coin

import (
	"encoding/binary"

	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type TxOutput struct {
	Amount uint64
	Owner  types.PublicKey
}

func (o TxOutput) Serialize() ([]byte, error) {
	if o.Owner.PublicKey == nil {
		return nil, NilPublicKeyError{}
	}

	key, err := o.Owner.Bytes()

	if err != nil {
		return nil, PublicKeySerializeError{Key: o.Owner, Previous: err}
	}

	buf := binary.BigEndian.AppendUint32(nil, uint32(len(key)))
	buf = append(buf, key...)
	buf = binary.BigEndian.AppendUint64(buf, o.Amount)

	return buf, nil
}

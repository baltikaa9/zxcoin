// Package coin определяет базовый примитив ценности в системе — выход транзакции (TxOutput).
package coin

import "crypto/ecdsa"

type TxOutput struct {
	Amount    int
	PublicKey *ecdsa.PublicKey
}

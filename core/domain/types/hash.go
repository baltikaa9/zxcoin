// Package types содержит общие примитивные типы доменной модели.
package types

import "encoding/hex"

type Hash [32]byte

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

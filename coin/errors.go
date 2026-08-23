package coin

import (
	"crypto/ecdsa"
	"fmt"
)

type PublicKeySerializeError struct {
	Key      *ecdsa.PublicKey
	Previous error
}

func (e *PublicKeySerializeError) Error() string {
	return fmt.Sprintf("ошибка при сериализации публичного ключа %v: %v", *e.Key, e.Previous)
}

type NilPublicKeyError struct{}

func (e *NilPublicKeyError) Error() string {
	return "публичный ключ не заполнен"
}

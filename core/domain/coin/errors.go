package coin

import (
	"fmt"

	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type PublicKeySerializeError struct {
	Key      types.PublicKey
	Previous error
}

func (e PublicKeySerializeError) Error() string {
	return fmt.Sprintf("ошибка при сериализации публичного ключа %v: %v", e.Key, e.Previous)
}

type NilPublicKeyError struct{}

func (e NilPublicKeyError) Error() string {
	return "публичный ключ не заполнен"
}

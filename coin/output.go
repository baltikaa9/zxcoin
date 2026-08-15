package coin

import "crypto/ecdsa"

type TxOutput struct {
	Amount    int
	PublicKey *ecdsa.PublicKey
}

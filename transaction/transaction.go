package transaction

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"zxcoin/coin"
	"zxcoin/utxo"
)

type TxInput struct {
	TxID      [32]byte
	OutIndex  int
	Signature Signature
}

type Signature struct {
	R *big.Int
	S *big.Int
}

type Transaction struct {
	Inputs  []TxInput
	Outputs []coin.TxOutput
}

func (t Transaction) Hash() [32]byte {
	tmp := t

	tmp.Inputs = make([]TxInput, len(t.Inputs))
	copy(tmp.Inputs, t.Inputs)

	for i := range tmp.Inputs {
		tmp.Inputs[i].Signature = Signature{}
	}

	data := fmt.Sprintf("%v", tmp)
	hash := sha256.Sum256([]byte(data))

	return hash
}

func (in *TxInput) Sign(privateKey *ecdsa.PrivateKey, hash [32]byte) error {
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return err
	}

	in.Signature = Signature{r, s}

	return nil
}

func (in *TxInput) Verify(publicKey *ecdsa.PublicKey, hash [32]byte) bool {
	return ecdsa.Verify(publicKey, hash[:], in.Signature.R, in.Signature.S)
}

func (t Transaction) Validate(utxoDB utxo.UTXODB) error {
	inputAmount := 0
	outputAmount := 0

	for _, input := range t.Inputs {
		key := utxo.UTXOKey{TxID: input.TxID, OutIndex: input.OutIndex}
		utxo, exists := utxoDB[key]

		if !exists {
			return &UTXONotFoundError{input.TxID, input.OutIndex}
		}

		if (input.Signature == Signature{}) || (!input.Verify(utxo.Output.PublicKey, t.Hash())) {
			return &InvalidSignatureError{input.TxID, input.OutIndex}
		}

		inputAmount += utxo.Output.Amount
	}

	for _, output := range t.Outputs {
		outputAmount += output.Amount
	}

	if outputAmount > inputAmount {
		return &NotEnoughMoneyError{inputAmount, outputAmount}
	}

	return nil
}

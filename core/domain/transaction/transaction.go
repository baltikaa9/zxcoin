// Package transaction определяет транзакции, их входы и выходы, а также логику их проверки: существование UTXO, корректность подписи и баланс сумм.
package transaction

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type TxInput struct {
	ID        utxo.UTXOID
	Signature Signature
}

type Signature struct {
	R *big.Int
	S *big.Int
}

func (s Signature) Marshal() []byte {
	raw := make([]byte, 64)

	s.R.FillBytes(raw[:32])
	s.S.FillBytes(raw[32:])

	return raw
}

func ParseSignature(raw []byte) (Signature, error) {
	if len(raw) != 64 {
		return Signature{}, fmt.Errorf("неверная длина подписи: %d", len(raw))
	}

	return Signature{
		R: new(big.Int).SetBytes(raw[:32]),
		S: new(big.Int).SetBytes(raw[32:]),
	}, nil
}

type Transaction struct {
	Inputs  []TxInput
	Outputs []coin.TxOutput
}

func NewTransaction(inputs []TxInput, outputs []coin.TxOutput, privateKey *ecdsa.PrivateKey) (Transaction, error) {
	t := Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
	hash, err := t.Hash()

	if err != nil {
		return Transaction{}, err
	}

	for i := range t.Inputs {
		t.Inputs[i].Sign(privateKey, hash)
	}

	return t, nil
}

func (t Transaction) Hash() (types.Hash, error) {
	data, err := t.serialize()

	if err != nil {
		return types.Hash{}, err
	}

	return sha256.Sum256(data), nil
}

func (in *TxInput) Sign(privateKey *ecdsa.PrivateKey, transactionHash types.Hash) error {
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, transactionHash[:])

	if err != nil {
		return err
	}

	in.Signature = Signature{r, s}

	return nil
}

func (in *TxInput) serialize() []byte {
	buf := in.ID.TxID[:]
	buf = binary.BigEndian.AppendUint64(buf, in.ID.OutIndex)

	return buf
}

func (in *TxInput) Verify(publicKey *ecdsa.PublicKey, hash types.Hash) bool {
	return ecdsa.Verify(publicKey, hash[:], in.Signature.R, in.Signature.S)
}

func (t Transaction) ValidateOutputs() error {
	hash, err := t.Hash()

	if err != nil {
		return err
	}

	if len(t.Outputs) == 0 {
		return EmptyOutputsError{TxID: hash}
	}

	for i, output := range t.Outputs {
		if output.Amount == 0 {
			return ZeroOutputError{TxID: hash, OutIndex: uint64(i)}
		}
	}

	return nil
}

func (t Transaction) VerifySignatures(spent []utxo.UTXO) error {
	hash, err := t.Hash()

	if err != nil {
		return err
	}

	emptySignature := Signature{}

	for i, input := range t.Inputs {
		if (input.Signature == emptySignature) || (!input.Verify(spent[i].Output.Owner, hash)) {
			return InvalidSignatureError{input.ID.TxID, input.ID.OutIndex}
		}
	}

	return nil
}

func (t Transaction) ValidateSum(spent []utxo.UTXO) error {
	inputAmount := uint64(0)
	outputAmount := uint64(0)

	for _, input := range spent {
		inputAmount += input.Output.Amount
	}

	for _, output := range t.Outputs {
		outputAmount += output.Amount
	}

	if outputAmount > inputAmount {
		return InsufficientFundsError{inputAmount, outputAmount}
	}

	return nil
}

func (t Transaction) serialize() ([]byte, error) {
	buf := binary.BigEndian.AppendUint32(nil, uint32(len(t.Inputs)))

	for _, input := range t.Inputs {
		buf = append(buf, input.serialize()...)
	}

	buf = binary.BigEndian.AppendUint32(buf, uint32(len(t.Outputs)))

	for _, output := range t.Outputs {
		data, err := output.Serialize()

		if err != nil {
			return nil, err
		}

		buf = append(buf, data...)
	}

	return buf, nil
}

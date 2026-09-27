// Package transaction определяет транзакции, их входы и выходы, а также логику их проверки: существование UTXO, корректность подписи и баланс сумм.
package transaction

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"zxcoin/coin"
	"zxcoin/utxo"
)

type TxInput struct {
	ID        utxo.UTXOID
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

func (t Transaction) Hash() ([32]byte, error) {
	data, err := t.serialize()

	if err != nil {
		return [32]byte{}, err
	}

	return sha256.Sum256(data), nil
}

func (in *TxInput) Sign(privateKey *ecdsa.PrivateKey, transactionHash [32]byte) error {
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

func (in *TxInput) Verify(publicKey *ecdsa.PublicKey, hash [32]byte) bool {
	return ecdsa.Verify(publicKey, hash[:], in.Signature.R, in.Signature.S)
}

func (t Transaction) Validate(utxoDB utxo.Repository) error {
	if err := t.validateInputs(utxoDB); err != nil {
		return err
	}

	if err := t.validateOutputs(); err != nil {
		return err
	}

	if err := t.validateSum(utxoDB); err != nil {
		return err
	}

	return nil
}

func (t Transaction) validateInputs(utxoDB utxo.Repository) error {
	hash, err := t.Hash()

	if err != nil {
		return err
	}

	emptySignature := Signature{}

	for _, input := range t.Inputs {
		utxo, exists, err := utxoDB.FindByID(input.ID)

		if err != nil {
			return err
		}

		if !exists {
			return UTXONotFoundError{input.ID.TxID, input.ID.OutIndex}
		}

		if (input.Signature == emptySignature) || (!input.Verify(utxo.Owner, hash)) {
			return InvalidSignatureError{input.ID.TxID, input.ID.OutIndex}
		}
	}

	return nil
}

func (t Transaction) validateOutputs() error {
	hash, err := t.Hash()

	if err != nil {
		return err
	}

	for i, output := range t.Outputs {
		if output.Amount == 0 {
			return ZeroOutputError{TxID: hash, OutIndex: uint64(i)}
		}
	}

	return nil
}

func (t Transaction) validateSum(utxoDB utxo.Repository) error {
	inputAmount := uint64(0)
	outputAmount := uint64(0)

	for _, input := range t.Inputs {
		utxo, _, _ := utxoDB.FindByID(input.ID)
		inputAmount += utxo.Amount
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

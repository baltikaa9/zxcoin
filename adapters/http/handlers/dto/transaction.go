package dto

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/baltikaa9/zxcoin/core/domain/transaction"
)

type Input struct {
	TxID      string `json:"tx_id"`
	OutIndex  uint64 `json:"out_index"`
	Signature string `json:"signature"`
}

type Output struct {
	Amount    uint64 `json:"amount"`
	PublicKey string `json:"public_key"`
}

type Transaction struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

func (d Transaction) Validate() error {
	if len(d.Inputs) == 0 {
		return errors.New("не заполнено поле inputs")
	}

	if len(d.Outputs) == 0 {
		return errors.New("не заполнено поле outputs")
	}

	for i, input := range d.Inputs {
		if err := validateHexField(
			fmt.Sprintf("inputs[%d].txID", i),
			input.TxID,
			32,
		); err != nil {
			return err
		}

		if err := validateHexField(
			fmt.Sprintf("inputs[%d].signature", i),
			input.Signature,
			64,
		); err != nil {
			return err
		}
	}

	for i, output := range d.Outputs {
		if output.Amount == 0 {
			return fmt.Errorf("outputs[%d].amount must be greater than zero", i)
		}

		if err := validateHexField(
			fmt.Sprintf("outputs[%d].publicKey", i),
			output.PublicKey,
			65,
		); err != nil {
			return err
		}

		if output.PublicKey[:2] != "04" {
			return fmt.Errorf("outputs[%d].publicKey must be uncompressed", i)
		}
	}

	return nil
}

func TransactionToDto(tx transaction.Transaction) (Transaction, error) {
	dto := Transaction{
		Inputs:  make([]Input, 0, len(tx.Inputs)),
		Outputs: make([]Output, 0, len(tx.Outputs)),
	}

	for _, input := range tx.Inputs {
		dto.Inputs = append(dto.Inputs, Input{
			TxID:      input.ID.TxID.String(),
			OutIndex:  input.ID.OutIndex,
			Signature: hex.EncodeToString(input.Signature.Marshal()),
		})
	}

	for _, output := range tx.Outputs {
		publicKey, err := output.Owner.Bytes()

		if err != nil {
			return Transaction{}, err
		}

		dto.Outputs = append(dto.Outputs, Output{
			Amount:    output.Amount,
			PublicKey: hex.EncodeToString(publicKey),
		})
	}

	return dto, nil
}

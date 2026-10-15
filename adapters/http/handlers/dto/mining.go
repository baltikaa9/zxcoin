package dto

import (
	"errors"
	"fmt"
)

type Mining struct {
	TransactionLimit uint64 `json:"transaction_limit"`
	Creator          string `json:"creator"`
}

func (d Mining) Validate() error {
	if d.TransactionLimit == 0 {
		return errors.New("transaction_limit должно быть больше 0")
	}

	if err := validateHexField("creator", d.Creator, 65); err != nil {
		return err
	}

	if d.Creator[:2] != "04" {
		return fmt.Errorf("creator publicKey must be uncompressed")
	}

	return nil
}

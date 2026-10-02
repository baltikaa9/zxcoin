package transaction

import (
	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type TransactionValidator struct {
	repo apputxo.Repository
}

func NewValidator(utxoRepo apputxo.Repository) *TransactionValidator {
	return &TransactionValidator{repo: utxoRepo}
}

func (v TransactionValidator) Validate(tx transaction.Transaction) error {
	spent, err := v.loadInputs(tx)

	if err != nil {
		return err
	}

	if err := tx.ValidateOutputs(); err != nil {
		return err
	}

	if err := tx.VerifySignatures(spent); err != nil {
		return err
	}

	return tx.ValidateSum(spent)
}

func (v TransactionValidator) loadInputs(tx transaction.Transaction) ([]utxo.UTXO, error) {
	spent := make([]utxo.UTXO, 0, len(tx.Inputs))

	for _, input := range tx.Inputs {
		u, exists, err := v.repo.FindByID(input.ID)

		if err != nil {
			return nil, err
		}

		if !exists {
			return nil, UTXONotFoundError{input.ID.TxID, input.ID.OutIndex}
		}

		spent = append(spent, u)
	}

	return spent, nil
}

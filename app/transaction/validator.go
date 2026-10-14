package transaction

import (
	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type Validator struct {
	repo apputxo.Repository
}

func NewValidator(utxoRepo apputxo.Repository) *Validator {
	return &Validator{repo: utxoRepo}
}

func (v Validator) Validate(tx transaction.Transaction) error {
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

func (v Validator) loadInputs(tx transaction.Transaction) ([]utxo.UTXO, error) {
	utxos := make([]utxo.UTXO, 0, len(tx.Inputs))
	spent := make(map[utxo.ID]struct{})

	for _, input := range tx.Inputs {
		if _, ok := spent[input.ID]; ok {
			return nil, DuplicateInputError{TxID: input.ID.TxID, OutIndex: input.ID.OutIndex}
		}

		u, exists, err := v.repo.FindByID(input.ID)

		if err != nil {
			return nil, err
		}

		if !exists {
			return nil, UTXONotFoundError{input.ID.TxID, input.ID.OutIndex}
		}

		utxos = append(utxos, u)
		spent[u.ID] = struct{}{}
	}

	return utxos, nil
}

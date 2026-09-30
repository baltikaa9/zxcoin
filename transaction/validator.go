package transaction

import "zxcoin/utxo"

type TransactionValidator struct {
	repo utxo.Repository
}

func NewValidator(utxoRepo utxo.Repository) *TransactionValidator {
	return &TransactionValidator{repo: utxoRepo}
}

func (v TransactionValidator) Validate(tx Transaction) error {
	spent, err := v.loadInputs(tx)

	if err != nil {
		return err
	}

	if err := tx.validateOutputs(); err != nil {
		return err
	}

	if err := v.verifySignatures(tx, spent); err != nil {
		return err
	}

	return v.validateSum(tx, spent)
}

func (v TransactionValidator) loadInputs(tx Transaction) ([]utxo.UTXO, error) {
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

func (v TransactionValidator) verifySignatures(tx Transaction, spent []utxo.UTXO) error {
	hash, err := tx.Hash()

	if err != nil {
		return err
	}

	emptySignature := Signature{}

	for i, input := range tx.Inputs {
		if (input.Signature == emptySignature) || (!input.Verify(spent[i].Output.Owner, hash)) {
			return InvalidSignatureError{input.ID.TxID, input.ID.OutIndex}
		}
	}

	return nil
}

func (v TransactionValidator) validateSum(tx Transaction, spent []utxo.UTXO) error {
	inputAmount := uint64(0)
	outputAmount := uint64(0)

	for _, input := range spent {
		inputAmount += input.Output.Amount
	}

	for _, output := range tx.Outputs {
		outputAmount += output.Amount
	}

	if outputAmount > inputAmount {
		return InsufficientFundsError{inputAmount, outputAmount}
	}

	return nil
}

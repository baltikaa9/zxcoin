package wallet

import (
	"crypto/ecdsa"
	"errors"
	"testing"
	"zxcoin/coin"
	"zxcoin/utxo"
)

func TestCreateTransaction_InsufficientFunds(t *testing.T) {
	myWallet := NewWallet()
	otherWallet := NewWallet()
	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: myWallet.PublicKey,
			},
		},
	}
	_, err := myWallet.CreateTransaction(otherWallet.PublicKey, 10, utxoDB)

	if _, ok := errors.AsType[*InsufficientFundsError](err); !ok {
		t.Fatalf("ожидалась InsufficientFundsError, получено: %v", err)
	}
}

func TestCreateTransaction_InsufficientFundsEmptyWallet(t *testing.T) {
	myWallet := NewWallet()
	otherWallet := NewWallet()
	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: otherWallet.PublicKey,
			},
		},
	}
	_, err := myWallet.CreateTransaction(otherWallet.PublicKey, 1, utxoDB)

	if _, ok := errors.AsType[*InsufficientFundsError](err); !ok {
		t.Fatalf("ожидалась InsufficientFundsError, получено: %v", err)
	}
}

func TestCreateTransaction_SuccessSingleInput(t *testing.T) {
	myWallet := NewWallet()
	otherWallet := NewWallet()
	key := utxo.UTXOKey{
		TxID:     [32]byte{},
		OutIndex: 0,
	}
	amount := 5
	utxoDB := utxo.UTXODB{
		key: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    amount,
				PublicKey: myWallet.PublicKey,
			},
		},
	}
	tx, err := myWallet.CreateTransaction(otherWallet.PublicKey, amount, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при создании транзакции: %v", err)
	}

	inputs := tx.Inputs
	outputs := tx.Outputs

	if len := len(inputs); len != 1 {
		t.Fatalf("неверное количество входов. Ожидалось 1, получено %v", len)
	}

	if len := len(outputs); len != 1 {
		t.Fatalf("неверное количество входов. Ожидалось 1, получено %v", len)
	}

	input := inputs[0]
	output := outputs[0]

	if input.TxID != key.TxID || input.OutIndex != key.OutIndex {
		t.Fatalf("неверно заполнен вход транзакции. Ожидалось %v - %v, получено %v - %v", key.TxID, key.OutIndex, input.TxID, input.OutIndex)
	}

	if output.Amount != amount {
		t.Fatalf("неверная сумма выхода транзакции. Ожидалось %v, получено %v", amount, output.Amount)
	}

	if !output.PublicKey.Equal(otherWallet.PublicKey) {
		t.Fatalf("неверный получатель транзакции. Ожидалось %v, получено %v", otherWallet.PublicKey, output.PublicKey)
	}

	hash := tx.Hash()

	if !ecdsa.Verify(myWallet.PublicKey, hash[:], input.Signature.R, input.Signature.S) {
		t.Fatalf("неверная подпись входа транзакции")
	}

	if !utxoDB[key].Reserved() {
		t.Fatalf("отсутсвует резервация utxo")
	}
}

func TestCreateTransaction_SuccessMultipleInput(t *testing.T) {
	myWallet := NewWallet()
	otherWallet := NewWallet()
	key0 := utxo.UTXOKey{
		TxID:     [32]byte{},
		OutIndex: 0,
	}
	key1 := utxo.UTXOKey{
		TxID:     [32]byte{},
		OutIndex: 1,
	}
	amount := 5
	utxoDB := utxo.UTXODB{
		key0: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    amount,
				PublicKey: myWallet.PublicKey,
			},
		},
		key1: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    amount,
				PublicKey: myWallet.PublicKey,
			},
		},
	}
	tx, err := myWallet.CreateTransaction(otherWallet.PublicKey, amount*2, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при создании транзакции: %v", err)
	}

	inputs := tx.Inputs
	outputs := tx.Outputs

	if len := len(inputs); len != 2 {
		t.Fatalf("неверное количество входов. Ожидалось 2, получено %v", len)
	}

	if len := len(outputs); len != 1 {
		t.Fatalf("неверное количество входов. Ожидалось 1, получено %v", len)
	}

	output := outputs[0]
	keysExist := map[utxo.UTXOKey]bool{}

	for _, input := range inputs {
		keysExist[utxo.UTXOKey{TxID: input.TxID, OutIndex: input.OutIndex}] = true
	}

	if !keysExist[key0] || !keysExist[key1] {
		t.Fatalf("не все ожидаемые входы присутствуют в транзакции: %v", keysExist)
	}

	if output.Amount != amount*2 {
		t.Fatalf("неверная сумма выхода транзакции. Ожидалось %v, получено %v", amount, output.Amount)
	}

	if !output.PublicKey.Equal(otherWallet.PublicKey) {
		t.Fatalf("неверный получатель транзакции. Ожидалось %v, получено %v", otherWallet.PublicKey, output.PublicKey)
	}

	hash := tx.Hash()

	for i, input := range inputs {
		if !ecdsa.Verify(myWallet.PublicKey, hash[:], input.Signature.R, input.Signature.S) {
			t.Fatalf("неверная подпись %v входа транзакции", i)
		}
	}

	if !utxoDB[key0].Reserved() {
		t.Fatalf("отсутсвует резервация 0 utxo")
	}

	if !utxoDB[key1].Reserved() {
		t.Fatalf("отсутсвует резервация 1 utxo")
	}
}

func TestCreateTransaction_SuccessChange(t *testing.T) {
	myWallet := NewWallet()
	otherWallet := NewWallet()
	key := utxo.UTXOKey{
		TxID:     [32]byte{},
		OutIndex: 0,
	}
	amount := 5
	payment := 4
	utxoDB := utxo.UTXODB{
		key: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    amount,
				PublicKey: myWallet.PublicKey,
			},
		},
	}
	tx, err := myWallet.CreateTransaction(otherWallet.PublicKey, payment, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при создании транзакции: %v", err)
	}

	inputs := tx.Inputs
	outputs := tx.Outputs

	if len := len(inputs); len != 1 {
		t.Fatalf("неверное количество входов. Ожидалось 2, получено %v", len)
	}

	if len := len(outputs); len != 2 {
		t.Fatalf("неверное количество входов. Ожидалось 1, получено %v", len)
	}

	input := inputs[0]
	output := outputs[0]
	change := outputs[1]

	if input.TxID != key.TxID || input.OutIndex != key.OutIndex {
		t.Fatalf("неверно заполнен 1 вход транзакции. Ожидалось %v - %v, получено %v - %v", key.TxID, key.OutIndex, input.TxID, input.OutIndex)
	}

	if output.Amount != payment {
		t.Fatalf("неверная сумма выхода транзакции. Ожидалось %v, получено %v", payment, output.Amount)
	}

	if change.Amount != amount-payment {
		t.Fatalf("неверная сумма сдачи транзакции. Ожидалось %v, получено %v", amount-payment, change.Amount)
	}

	if !output.PublicKey.Equal(otherWallet.PublicKey) {
		t.Fatalf("неверный получатель транзакции. Ожидалось %v, получено %v", otherWallet.PublicKey, output.PublicKey)
	}

	if !change.PublicKey.Equal(myWallet.PublicKey) {
		t.Fatalf("неверный получатель сдачи. Ожидалось %v, получено %v", myWallet.PublicKey, change.PublicKey)
	}

	hash := tx.Hash()

	if !ecdsa.Verify(myWallet.PublicKey, hash[:], input.Signature.R, input.Signature.S) {
		t.Fatalf("неверная подпись входа транзакции")
	}

	if !utxoDB[key].Reserved() {
		t.Fatalf("отсутсвует резервация utxo")
	}
}

func TestCreateTransaction_SuccessMultipleInputChange(t *testing.T) {
	myWallet := NewWallet()
	otherWallet := NewWallet()
	key0 := utxo.UTXOKey{
		TxID:     [32]byte{},
		OutIndex: 0,
	}
	key1 := utxo.UTXOKey{
		TxID:     [32]byte{},
		OutIndex: 1,
	}
	amount := 5
	payment := 4
	utxoDB := utxo.UTXODB{
		key0: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    amount,
				PublicKey: myWallet.PublicKey,
			},
		},
		key1: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    amount,
				PublicKey: myWallet.PublicKey,
			},
		},
	}
	tx, err := myWallet.CreateTransaction(otherWallet.PublicKey, payment*2, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при создании транзакции: %v", err)
	}

	inputs := tx.Inputs
	outputs := tx.Outputs

	if len := len(inputs); len != 2 {
		t.Fatalf("неверное количество входов. Ожидалось 2, получено %v", len)
	}

	if len := len(outputs); len != 2 {
		t.Fatalf("неверное количество входов. Ожидалось 1, получено %v", len)
	}

	output := outputs[0]
	change := outputs[1]
	keysExist := map[utxo.UTXOKey]bool{}

	for _, input := range inputs {
		keysExist[utxo.UTXOKey{TxID: input.TxID, OutIndex: input.OutIndex}] = true
	}

	if !keysExist[key0] || !keysExist[key1] {
		t.Fatalf("не все ожидаемые входы присутствуют в транзакции: %v", keysExist)
	}

	if output.Amount != payment*2 {
		t.Fatalf("неверная сумма выхода транзакции. Ожидалось %v, получено %v", payment*2, output.Amount)
	}

	if change.Amount != amount*2-payment*2 {
		t.Fatalf("неверная сумма выхода транзакции. Ожидалось %v, получено %v", amount*2-payment*2, change.Amount)
	}

	if !output.PublicKey.Equal(otherWallet.PublicKey) {
		t.Fatalf("неверный получатель транзакции. Ожидалось %v, получено %v", otherWallet.PublicKey, output.PublicKey)
	}

	if !change.PublicKey.Equal(myWallet.PublicKey) {
		t.Fatalf("неверный получатель сдачи. Ожидалось %v, получено %v", myWallet.PublicKey, change.PublicKey)
	}

	hash := tx.Hash()

	for i, input := range inputs {
		if !ecdsa.Verify(myWallet.PublicKey, hash[:], input.Signature.R, input.Signature.S) {
			t.Fatalf("неверная подпись %v входа транзакции", i)
		}
	}

	if !utxoDB[key0].Reserved() {
		t.Fatalf("отсутсвует резервация 0 utxo")
	}

	if !utxoDB[key1].Reserved() {
		t.Fatalf("отсутсвует резервация 1 utxo")
	}
}

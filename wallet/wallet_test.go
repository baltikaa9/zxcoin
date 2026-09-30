package wallet

import (
	"crypto/ecdsa"
	"errors"
	"testing"
	"zxcoin/coin"
	"zxcoin/testutil"
	"zxcoin/types"
	"zxcoin/utxo"
	"zxcoin/utxo/inmemory"
)

func TestCreateTransaction_InsufficientFunds(t *testing.T) {
	repo := inmemory.NewRepository()
	tracker := NewReservationTracker()

	myWallet := NewWallet(tracker, repo)
	otherWallet := NewWallet(tracker, repo)

	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, myWallet.PublicKey, repo)
	_, err := myWallet.CreateTransaction(otherWallet.PublicKey, amount*2)

	if _, ok := errors.AsType[InsufficientFundsError](err); !ok {
		t.Fatalf("ожидалась InsufficientFundsError, получено: %v", err)
	}
}

func TestCreateTransaction_InsufficientFundsEmptyWallet(t *testing.T) {
	repo := inmemory.NewRepository()
	tracker := NewReservationTracker()

	myWallet := NewWallet(tracker, repo)
	otherWallet := NewWallet(tracker, repo)

	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, otherWallet.PublicKey, repo)
	_, err := myWallet.CreateTransaction(otherWallet.PublicKey, 1)

	if _, ok := errors.AsType[InsufficientFundsError](err); !ok {
		t.Fatalf("ожидалась InsufficientFundsError, получено: %v", err)
	}
}

func TestCreateTransaction_SuccessSingleInput(t *testing.T) {
	repo := inmemory.NewRepository()
	tracker := NewReservationTracker()

	myWallet := NewWallet(tracker, repo)
	otherWallet := NewWallet(tracker, repo)

	id := utxo.UTXOID{
		TxID:     types.Hash{},
		OutIndex: 0,
	}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, myWallet.PublicKey, repo)
	tx, err := myWallet.CreateTransaction(otherWallet.PublicKey, amount)

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

	if input.ID != id {
		t.Fatalf("неверно заполнен вход транзакции. Ожидалось %v - %v, получено %v - %v", id.TxID, id.OutIndex, input.ID.TxID, input.ID.OutIndex)
	}

	if output.Amount != amount {
		t.Fatalf("неверная сумма выхода транзакции. Ожидалось %v, получено %v", amount, output.Amount)
	}

	if !output.Owner.Equal(otherWallet.PublicKey) {
		t.Fatalf("неверный получатель транзакции. Ожидалось %v, получено %v", otherWallet.PublicKey, output.Owner)
	}

	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if !ecdsa.Verify(myWallet.PublicKey, hash[:], input.Signature.R, input.Signature.S) {
		t.Fatalf("неверная подпись входа транзакции")
	}

	if !tracker.IsReserved(id) {
		t.Fatalf("отсутсвует резервация utxo")
	}
}

func TestCreateTransaction_SuccessMultipleInput(t *testing.T) {
	repo := inmemory.NewRepository()
	tracker := NewReservationTracker()

	myWallet := NewWallet(tracker, repo)
	otherWallet := NewWallet(tracker, repo)

	id0 := utxo.UTXOID{
		TxID:     types.Hash{},
		OutIndex: 0,
	}
	id1 := utxo.UTXOID{
		TxID:     types.Hash{},
		OutIndex: 1,
	}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, myWallet.PublicKey, repo)
	if err := repo.Save(utxo.UTXO{
		ID: id1,
		Output: coin.TxOutput{
			Amount: amount,
			Owner:  myWallet.PublicKey,
		},
	}); err != nil {
		t.Fatalf("не удалось сохранить UTXO: %v", err)
	}
	tx, err := myWallet.CreateTransaction(otherWallet.PublicKey, amount*2)

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
	keysExist := map[utxo.UTXOID]bool{}

	for _, input := range inputs {
		keysExist[input.ID] = true

	}

	if !keysExist[id0] || !keysExist[id1] {
		t.Fatalf("не все ожидаемые входы присутствуют в транзакции: %v", keysExist)
	}

	if output.Amount != amount*2 {
		t.Fatalf("неверная сумма выхода транзакции. Ожидалось %v, получено %v", amount, output.Amount)
	}

	if !output.Owner.Equal(otherWallet.PublicKey) {
		t.Fatalf("неверный получатель транзакции. Ожидалось %v, получено %v", otherWallet.PublicKey, output.Owner)
	}

	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	for i, input := range inputs {
		if !ecdsa.Verify(myWallet.PublicKey, hash[:], input.Signature.R, input.Signature.S) {
			t.Fatalf("неверная подпись %v входа транзакции", i)
		}
	}

	if !tracker.IsReserved(id0) {
		t.Fatalf("отсутствует резервация 0 utxo")
	}

	if !tracker.IsReserved(id1) {
		t.Fatalf("отсутствует резервация 1 utxo")
	}
}

func TestCreateTransaction_SuccessChange(t *testing.T) {
	repo := inmemory.NewRepository()
	tracker := NewReservationTracker()

	myWallet := NewWallet(tracker, repo)
	otherWallet := NewWallet(tracker, repo)

	id := utxo.UTXOID{
		TxID:     types.Hash{},
		OutIndex: 0,
	}
	amount := uint64(5)
	payment := uint64(4)
	testutil.GenerateSingleUtxo(t, amount, myWallet.PublicKey, repo)
	tx, err := myWallet.CreateTransaction(otherWallet.PublicKey, payment)

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

	if input.ID != id {
		t.Fatalf("неверно заполнен 1 вход транзакции. Ожидалось %v - %v, получено %v - %v", id.TxID, id.OutIndex, input.ID.TxID, input.ID.OutIndex)
	}

	if output.Amount != payment {
		t.Fatalf("неверная сумма выхода транзакции. Ожидалось %v, получено %v", payment, output.Amount)
	}

	if change.Amount != amount-payment {
		t.Fatalf("неверная сумма сдачи транзакции. Ожидалось %v, получено %v", amount-payment, change.Amount)
	}

	if !output.Owner.Equal(otherWallet.PublicKey) {
		t.Fatalf("неверный получатель транзакции. Ожидалось %v, получено %v", otherWallet.PublicKey, output.Owner)
	}

	if !change.Owner.Equal(myWallet.PublicKey) {
		t.Fatalf("неверный получатель сдачи. Ожидалось %v, получено %v", myWallet.PublicKey, change.Owner)
	}

	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if !ecdsa.Verify(myWallet.PublicKey, hash[:], input.Signature.R, input.Signature.S) {
		t.Fatalf("неверная подпись входа транзакции")
	}

	if !tracker.IsReserved(id) {
		t.Fatalf("отсутсвует резервация utxo")
	}
}

func TestCreateTransaction_SuccessMultipleInputChange(t *testing.T) {
	repo := inmemory.NewRepository()
	tracker := NewReservationTracker()

	myWallet := NewWallet(tracker, repo)
	otherWallet := NewWallet(tracker, repo)

	id0 := utxo.UTXOID{
		TxID:     types.Hash{},
		OutIndex: 0,
	}
	id1 := utxo.UTXOID{
		TxID:     types.Hash{},
		OutIndex: 1,
	}
	amount := uint64(5)
	payment := uint64(4)
	testutil.GenerateSingleUtxo(t, amount, myWallet.PublicKey, repo)
	if err := repo.Save(utxo.UTXO{
		ID: id1,
		Output: coin.TxOutput{
			Amount: amount,
			Owner:  myWallet.PublicKey,
		},
	}); err != nil {
		t.Fatalf("не удалось сохранить UTXO: %v", err)
	}
	tx, err := myWallet.CreateTransaction(otherWallet.PublicKey, payment*2)

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
	keysExist := map[utxo.UTXOID]bool{}

	for _, input := range inputs {
		keysExist[input.ID] = true
	}

	if !keysExist[id0] || !keysExist[id1] {
		t.Fatalf("не все ожидаемые входы присутствуют в транзакции: %v", keysExist)
	}

	if output.Amount != payment*2 {
		t.Fatalf("неверная сумма выхода транзакции. Ожидалось %v, получено %v", payment*2, output.Amount)
	}

	if change.Amount != amount*2-payment*2 {
		t.Fatalf("неверная сумма выхода транзакции. Ожидалось %v, получено %v", amount*2-payment*2, change.Amount)
	}

	if !output.Owner.Equal(otherWallet.PublicKey) {
		t.Fatalf("неверный получатель транзакции. Ожидалось %v, получено %v", otherWallet.PublicKey, output.Owner)
	}

	if !change.Owner.Equal(myWallet.PublicKey) {
		t.Fatalf("неверный получатель сдачи. Ожидалось %v, получено %v", myWallet.PublicKey, change.Owner)
	}

	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	for i, input := range inputs {
		if !ecdsa.Verify(myWallet.PublicKey, hash[:], input.Signature.R, input.Signature.S) {
			t.Fatalf("неверная подпись %v входа транзакции", i)
		}
	}

	if !tracker.IsReserved(id0) {
		t.Fatalf("отсутсвует резервация 0 utxo")
	}

	if !tracker.IsReserved(id1) {
		t.Fatalf("отсутсвует резервация 1 utxo")
	}
}

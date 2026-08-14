package main

import (
	"errors"
	"testing"
)

func TestAddBlock_DoubleSpendInBlock(t *testing.T) {
	wallet := NewWallet()
	other := NewWallet()

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: UTXOEntry{TxOutput{5, wallet.PublicKey}, false},
	}
	mempool := NewMempool()

	t1, err := wallet.CreateTransaction(other.PublicKey, 5, utxoDB)

	if err != nil {
		t.Errorf("ошибка при создании честной транзакции: %v", err)
	}

	t2 := Transaction{[]TxInput{{TxID: [32]byte{}, OutIndex: 0}}, []TxOutput{{5, wallet.PublicKey}}}
	t2.Inputs[0].Sign(wallet.PrivateKey, t2.Hash())

	err = mempool.Add(t1, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции в пул: %v", err)
	}

	err = mempool.Add(t2, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции в пул: %v", err)
	}

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	_, err = bc.MineAndAddBlock(mempool, utxoDB, 2, wallet.PublicKey)

	if _, ok := errors.AsType[*DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_DoubleSpendInTransaction(t *testing.T) {
	wallet := NewWallet()
	other := NewWallet()

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: UTXOEntry{TxOutput{5, wallet.PublicKey}, false},
	}
	mempool := NewMempool()

	t1 := Transaction{
		Inputs: []TxInput{
			{TxID: [32]byte{}, OutIndex: 0},
			{TxID: [32]byte{}, OutIndex: 0},
		},
		Outputs: []TxOutput{
			{5, other.PublicKey},
		},
	}

	hash := t1.Hash()

	for i := range t1.Inputs {
		t1.Inputs[i].Sign(wallet.PrivateKey, hash)
	}

	err := mempool.Add(t1, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции: %v", err)
	}

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	_, err = bc.MineAndAddBlock(mempool, utxoDB, 1, wallet.PublicKey)

	if _, ok := errors.AsType[*DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_InvalidNonce(t *testing.T) {
	wallet := NewWallet()
	other := NewWallet()

	utxoDB := UTXODB{
		UTXOKey{[32]byte{}, 0}: UTXOEntry{TxOutput{5, wallet.PublicKey}, false},
	}
	mempool := NewMempool()

	tx := Transaction{
		Inputs: []TxInput{
			{TxID: [32]byte{}, OutIndex: 0},
		},
		Outputs: []TxOutput{
			{5, other.PublicKey},
		},
	}
	tx.Inputs[0].Sign(wallet.PrivateKey, tx.Hash())

	err := mempool.Add(tx, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции: %v", err)
	}

	bc := Blockchain{CurrentDifficulty: 2, CurrentAward: 1}
	block := bc.NewBlock(mempool.GetPending(1), wallet.PublicKey)

	err = bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidNonceError](err); !ok {
		t.Fatalf("ожидалась InvalidNonceError, получено: %v", err)
	}
}

func TestAddBlock_InvalidPrevHash(t *testing.T) {
	wallet := NewWallet()
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := UTXODB{}

	genesisBlock := bc.NewBlock([]Transaction{}, wallet.PublicKey)
	genesisBlock.Mine()

	err := bc.AddBlock(genesisBlock, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении первого блока: %v", err)
	}

	block := Block{
		BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		[]Transaction{},
		bc.CurrentDifficulty,
	}
	block.Mine()

	err = bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidPrevHashError](err); !ok {
		t.Fatalf("ожидалась InvalidPrevHashError, получено: %v", err)
	}
}

func TestAddBlock_InvalidMerkleRootHash(t *testing.T) {
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := UTXODB{}

	block := Block{
		BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{1},
			Timestamp: 0,
		},
		[]Transaction{},
		bc.CurrentDifficulty,
	}
	block.Mine()

	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidMerkleRootError](err); !ok {
		t.Fatalf("ожидалась InvalidMerkleRootError, получено: %v", err)
	}
}

func TestAddBlock_MoreOneCoinbase(t *testing.T) {
	wallet := NewWallet()
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := UTXODB{}

	block := Block{
		BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		[]Transaction{
			{Outputs: []TxOutput{{bc.CurrentAward, wallet.PublicKey}}},
			{Outputs: []TxOutput{{bc.CurrentAward, wallet.PublicKey}}},
		},
		bc.CurrentDifficulty,
	}
	block.calculateRootHash()
	block.Mine()

	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*MoreOneCoinbaseError](err); !ok {
		t.Fatalf("ожидалась MoreOneCoinbaseError, получено: %v", err)
	}
}

func TestAddBlock_CoinbaseExisted(t *testing.T) {
	wallet := NewWallet()
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := UTXODB{}

	block := bc.NewBlock([]Transaction{}, wallet.PublicKey)

	if len(block.Transactions) < 1 {
		t.Fatalf("отсутствует coinbase-транзакция")
	}

	tx := block.Transactions[0]

	if len(tx.Inputs) > 0 || len(tx.Outputs) != 1 {
		t.Fatalf("некорректная coinbase-транзакция")
	}

	output := tx.Outputs[0]

	if output.PublicKey != wallet.PublicKey {
		t.Fatalf("неверный получатель coinbase-транзакции")
	}

	if output.Amount != bc.CurrentAward {
		t.Fatalf("неверный размер coinbase-транзакции")
	}

	block.Mine()
	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*MoreOneCoinbaseError](err); ok {
		t.Fatalf("несколько coinbase-транзакций")
	}

	if len(utxoDB) == 0 {
		t.Fatalf("coinbase-транзакция не добавилась в utxodb")
	}

	utxo, existed := utxoDB[UTXOKey{tx.Hash(), 0}]

	if !existed {
		t.Fatalf("некорректный ключ coinbase-транзакции в utxodb")
	}

	if utxo.Reserved {
		t.Fatalf("некорректная резервация coinbase-транзакции в utxodb")
	}

	if utxo.Output.Amount != bc.CurrentAward {
		t.Fatalf("некорректная сумма coinbase-транзакции в utxodb")
	}

	if utxo.Output.PublicKey != wallet.PublicKey {
		t.Fatalf("некорректный получатель coinbase-транзакции в utxodb")
	}
}

func TestAddBlock_Success(t *testing.T) {
	wallet := NewWallet()
	other := NewWallet()

	bc := Blockchain{CurrentDifficulty: 2, CurrentAward: 5}
	utxoDB := UTXODB{}
	mempool := NewMempool()

	genesisBlock := bc.NewBlock([]Transaction{}, wallet.PublicKey)
	genesisBlock.Mine()

	err := bc.AddBlock(genesisBlock, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при добавлении первого блока: %v", err)
	}

	tx, err := wallet.CreateTransaction(other.PublicKey, 5, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при создании транзакции: %v", err)
	}

	err = mempool.Add(tx, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при добавлении транзакции: %v", err)
	}

	_, err = bc.MineAndAddBlock(mempool, utxoDB, 1, wallet.PublicKey)

	if err != nil {
		t.Fatalf("ошибка при создании и добавлении блока: %v", err)
	}

	utxo, existed := utxoDB[UTXOKey{tx.Hash(), 0}]

	if !existed {
		t.Fatalf("некорректный ключ транзакции в utxodb")
	}

	if utxo.Reserved {
		t.Fatalf("некорректная резервация транзакции в utxodb")
	}

	if utxo.Output.Amount != bc.CurrentAward {
		t.Fatalf("некорректная сумма транзакции в utxodb")
	}

	if utxo.Output.PublicKey != other.PublicKey {
		t.Fatalf("некорректный получатель транзакции в utxodb")
	}
}

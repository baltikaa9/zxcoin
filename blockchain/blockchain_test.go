package blockchain

import (
	"crypto/ecdsa"
	"errors"
	"testing"
	"zxcoin/block"
	"zxcoin/coin"
	"zxcoin/mempool"
	"zxcoin/merkle"
	"zxcoin/testutil"
	"zxcoin/transaction"
	"zxcoin/utxo"
)

func mineGenesisBlock(t *testing.T, bc *Blockchain, utxoDB utxo.UTXODB) block.Block {
	t.Helper()
	genesisBlock := block.Block{
		Header:       block.BlockHeader{},
		Transactions: []transaction.Transaction{},
		Difficulty:   bc.currentDifficulty,
	}
	genesisBlock.CalculateRootHash()
	genesisBlock.Mine()

	if err := bc.AddBlock(genesisBlock, utxoDB); err != nil {
		t.Fatalf("ошибка при добавлении генезис-блока: %v", err)
	}

	return genesisBlock
}

func assertUTXO(t *testing.T, utxoDB utxo.UTXODB, key utxo.UTXOKey, expectedAmount int, expectedOwner *ecdsa.PublicKey) {
	t.Helper()
	entry, existed := utxoDB[key]

	if !existed {
		t.Fatalf("UTXO не найден: %v", key)
	}

	if entry.Reserved() {
		t.Fatalf("UTXO некорректно зарезервирован: %v", key)
	}

	if entry.Output.Amount != expectedAmount {
		t.Fatalf("неверная сумма UTXO. Ожидалось %v, получено %v", expectedAmount, entry.Output.Amount)
	}

	if !entry.Output.PublicKey.Equal(expectedOwner) {
		t.Fatalf("неверный владелец UTXO")
	}
}

func TestAddBlock_DoubleSpendInBlock(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := 5
	utxoDB := testutil.GenerateSingleUtxo(t, amount, publicKey)

	t1 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: amount, PublicKey: publicKey}},
	}
	t1.Inputs[0].Sign(privateKey, t1.Hash())

	t2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: amount, PublicKey: publicKey}},
	}
	t2.Inputs[0].Sign(privateKey, t2.Hash())

	bc := Blockchain{currentDifficulty: 1, currentAward: 1}

	block := bc.newBlock([]transaction.Transaction{t1, t2}, publicKey)
	block.Mine()
	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_DoubleSpendInTransaction(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := 5
	utxoDB := testutil.GenerateSingleUtxo(t, amount, publicKey)

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{TxID: [32]byte{}, OutIndex: 0},
			{TxID: [32]byte{}, OutIndex: 0},
		},
		Outputs: []coin.TxOutput{
			{Amount: amount, PublicKey: publicKey},
		},
	}

	hash := tx.Hash()

	for i := range tx.Inputs {
		tx.Inputs[i].Sign(privateKey, hash)
	}

	bc := Blockchain{currentDifficulty: 1, currentAward: 1}

	block := bc.newBlock([]transaction.Transaction{tx}, publicKey)
	block.Mine()
	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_InvalidNonce(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := 5
	utxoDB := testutil.GenerateSingleUtxo(t, amount, publicKey)

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{TxID: [32]byte{}, OutIndex: 0},
		},
		Outputs: []coin.TxOutput{
			{Amount: amount, PublicKey: publicKey},
		},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())

	bc := Blockchain{currentDifficulty: 2, currentAward: 1}
	block := bc.newBlock([]transaction.Transaction{tx}, publicKey)

	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidNonceError](err); !ok {
		t.Fatalf("ожидалась InvalidNonceError, получено: %v", err)
	}
}

func TestAddBlock_InvalidPrevHash(t *testing.T) {
	bc := Blockchain{currentDifficulty: 1, currentAward: 1}
	utxoDB := utxo.UTXODB{}
	mineGenesisBlock(t, &bc, utxoDB)
	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
		Difficulty:   bc.currentDifficulty,
	}
	block.CalculateRootHash()
	block.Mine()
	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidPrevHashError](err); !ok {
		t.Fatalf("ожидалась InvalidPrevHashError, получено: %v", err)
	}
}

func TestAddBlock_InvalidMerkleRootHash(t *testing.T) {
	bc := Blockchain{currentDifficulty: 1, currentAward: 1}
	utxoDB := utxo.UTXODB{}
	genesisBlock := mineGenesisBlock(t, &bc, utxoDB)
	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  genesisBlock.Header.Hash(),
			RootHash:  [32]byte{1},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
		Difficulty:   bc.currentDifficulty,
	}
	block.Mine()
	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidMerkleRootError](err); !ok {
		t.Fatalf("ожидалась InvalidMerkleRootError, получено: %v", err)
	}
}

func TestAddBlock_MoreOneCoinbase(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	bc := Blockchain{currentDifficulty: 1, currentAward: 1}
	utxoDB := utxo.UTXODB{}
	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{
			{Outputs: []coin.TxOutput{{Amount: bc.currentAward, PublicKey: publicKey}}},
			{Outputs: []coin.TxOutput{{Amount: bc.currentAward, PublicKey: publicKey}}},
		},
		Difficulty: bc.currentDifficulty,
	}

	block.CalculateRootHash()
	block.Mine()
	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*MoreOneCoinbaseError](err); !ok {
		t.Fatalf("ожидалась MoreOneCoinbaseError, получено: %v", err)
	}
}

func TestNewBlock_CoinbaseExisted(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	bc := Blockchain{currentDifficulty: 1, currentAward: 1}
	block := bc.newBlock([]transaction.Transaction{}, publicKey)

	if len(block.Transactions) < 1 {
		t.Fatalf("отсутствует coinbase-транзакция")
	}

	tx := block.Transactions[0]

	if len(tx.Inputs) > 0 || len(tx.Outputs) != 1 {
		t.Fatalf("некорректная coinbase-транзакция")
	}

	output := tx.Outputs[0]

	if !output.PublicKey.Equal(publicKey) {
		t.Fatalf("неверный получатель coinbase-транзакции")
	}

	if output.Amount != bc.currentAward {
		t.Fatalf("неверный размер coinbase-транзакции")
	}
}

func TestNewBlock_ValidPrevHash(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	bc := Blockchain{currentDifficulty: 1, currentAward: 1}
	utxoDB := utxo.UTXODB{}

	genesisBlock := mineGenesisBlock(t, &bc, utxoDB)
	block := bc.newBlock([]transaction.Transaction{}, publicKey)
	expectedHash := genesisBlock.Header.Hash()

	if block.Header.PrevHash != expectedHash {
		t.Fatalf("PrevHash не совпадает. Ожидалось: %v, получено: %v", expectedHash, block.Header.PrevHash)
	}
}

func TestNewBlock_ValidRootHash(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	bc := Blockchain{currentDifficulty: 1, currentAward: 1}
	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{1}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 3, PublicKey: publicKey}},
	}
	coinbaseTx := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: bc.currentAward, PublicKey: publicKey}},
	}

	block := bc.newBlock([]transaction.Transaction{tx}, publicKey)
	expectedHash := merkle.BuildMerkleTree([]transaction.Transaction{tx, coinbaseTx}).Hash

	if block.Header.RootHash != expectedHash {
		t.Fatalf("RootHash не совпадает. Ожидалось: %v, получено: %v", expectedHash, block.Header.RootHash)
	}
}

func TestAddBlock_CoinbaseExisted(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	bc := Blockchain{currentDifficulty: 1, currentAward: 1}
	utxoDB := utxo.UTXODB{}
	tx := transaction.Transaction{Outputs: []coin.TxOutput{{Amount: bc.currentAward, PublicKey: publicKey}}}
	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{tx},
		Difficulty:   bc.currentDifficulty,
	}
	block.CalculateRootHash()
	block.Mine()

	err := bc.AddBlock(block, utxoDB)

	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(utxoDB) == 0 {
		t.Fatalf("coinbase-транзакция не добавилась в utxodb")
	}

	assertUTXO(t, utxoDB, utxo.UTXOKey{TxID: tx.Hash(), OutIndex: 0}, bc.currentAward, publicKey)
}

func TestMineAndAddBlock_Success(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	_, otherPublicKey := testutil.GenerateKeyPair(t)
	bc := Blockchain{currentDifficulty: 2, currentAward: 5}
	utxoDB := utxo.UTXODB{}
	mempool := mempool.NewMempool()

	genesisBlock := bc.newBlock([]transaction.Transaction{}, publicKey)
	genesisBlock.Mine()
	err := bc.AddBlock(genesisBlock, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при добавлении первого блока: %v", err)
	}

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{TxID: genesisBlock.Transactions[0].Hash(), OutIndex: 0},
		},
		Outputs: []coin.TxOutput{
			{Amount: bc.currentAward, PublicKey: otherPublicKey},
		},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())

	err = mempool.Add(tx, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при добавлении транзакции: %v", err)
	}

	_, err = bc.MineAndAddBlock(mempool, utxoDB, 1, publicKey)

	if err != nil {
		t.Fatalf("ошибка при создании и добавлении блока: %v", err)
	}

	assertUTXO(t, utxoDB, utxo.UTXOKey{TxID: tx.Hash(), OutIndex: 0}, bc.currentAward, otherPublicKey)
}

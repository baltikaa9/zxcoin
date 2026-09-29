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
	"zxcoin/utxo/inmemory"
)

func mineGenesisBlock(t *testing.T, bc *Blockchain, repo utxo.Repository) block.Block {
	t.Helper()
	genesisBlock := block.Block{
		Header:       block.BlockHeader{},
		Transactions: []transaction.Transaction{},
		Difficulty:   bc.currentDifficulty,
	}
	err := genesisBlock.CalculateRootHash()

	if err != nil {
		t.Fatal(err)
	}

	genesisBlock.Mine()

	if err := bc.AddBlock(genesisBlock, repo); err != nil {
		t.Fatalf("ошибка при добавлении генезис-блока: %v", err)
	}

	return genesisBlock
}

func assertUTXO(t *testing.T, repo utxo.Repository, id utxo.UTXOID, expectedAmount uint64, expectedOwner *ecdsa.PublicKey) {
	t.Helper()
	u, existed, err := repo.FindByID(id)

	if err != nil {
		t.Fatalf("не удалось получить UTXO: %v", err)
	}

	if !existed {
		t.Fatalf("UTXO не найден: %v", id)
	}

	if u.Amount != expectedAmount {
		t.Fatalf("неверная сумма UTXO. Ожидалось %v, получено %v", expectedAmount, u.Amount)
	}

	if !u.Owner.Equal(expectedOwner) {
		t.Fatalf("неверный владелец UTXO")
	}
}

func TestAddBlock_DoubleSpendInBlock(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := transaction.TransactionValidator{UtxoRepo: repo}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	t1 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: amount, PublicKey: publicKey}},
	}
	t1Hash, err := t1.Hash()

	if err != nil {
		t.Fatal(err)
	}

	t1.Inputs[0].Sign(privateKey, t1Hash)

	t2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: amount, PublicKey: publicKey}},
	}
	t2Hash, err := t2.Hash()

	if err != nil {
		t.Fatal(err)
	}

	t2.Inputs[0].Sign(privateKey, t2Hash)

	bc := NewBlockchain(1, 1, &validator)

	block, err := bc.newBlock([]transaction.Transaction{t1, t2}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	block.Mine()
	err = bc.AddBlock(block, repo)

	if _, ok := errors.AsType[DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_DoubleSpendInTransaction(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := transaction.TransactionValidator{UtxoRepo: repo}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{ID: utxo.UTXOID{}},
			{ID: utxo.UTXOID{}},
		},
		Outputs: []coin.TxOutput{
			{Amount: amount, PublicKey: publicKey},
		},
	}

	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	for i := range tx.Inputs {
		tx.Inputs[i].Sign(privateKey, hash)
	}

	bc := NewBlockchain(1, 1, &validator)

	block, err := bc.newBlock([]transaction.Transaction{tx}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	block.Mine()
	err = bc.AddBlock(block, repo)

	if _, ok := errors.AsType[DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_InvalidNonce(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := transaction.TransactionValidator{UtxoRepo: repo}
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, repo)

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{ID: utxo.UTXOID{}},
		},
		Outputs: []coin.TxOutput{
			{Amount: amount, PublicKey: publicKey},
		},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, hash)

	bc := NewBlockchain(2, 1, &validator)
	block, err := bc.newBlock([]transaction.Transaction{tx}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	err = bc.AddBlock(block, repo)

	if _, ok := errors.AsType[InvalidNonceError](err); !ok {
		t.Fatalf("ожидалась InvalidNonceError, получено: %v", err)
	}
}

func TestAddBlock_InvalidPrevHash(t *testing.T) {
	repo := inmemory.NewRepository()
	validator := transaction.TransactionValidator{UtxoRepo: repo}
	bc := NewBlockchain(1, 1, &validator)
	mineGenesisBlock(t, &bc, repo)
	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
		Difficulty:   bc.currentDifficulty,
	}
	err := block.CalculateRootHash()

	if err != nil {
		t.Fatal(err)
	}

	block.Mine()
	err = bc.AddBlock(block, repo)

	if _, ok := errors.AsType[InvalidPrevHashError](err); !ok {
		t.Fatalf("ожидалась InvalidPrevHashError, получено: %v", err)
	}
}

func TestAddBlock_InvalidMerkleRootHash(t *testing.T) {
	repo := inmemory.NewRepository()
	validator := transaction.TransactionValidator{UtxoRepo: repo}
	bc := NewBlockchain(1, 1, &validator)
	genesisBlock := mineGenesisBlock(t, &bc, repo)
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
	err := bc.AddBlock(block, repo)

	if _, ok := errors.AsType[InvalidMerkleRootError](err); !ok {
		t.Fatalf("ожидалась InvalidMerkleRootError, получено: %v", err)
	}
}

func TestAddBlock_MoreOneCoinbase(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := transaction.TransactionValidator{UtxoRepo: repo}
	bc := NewBlockchain(1, 1, &validator)
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

	err := block.CalculateRootHash()

	if err != nil {
		t.Fatal(err)
	}

	block.Mine()
	err = bc.AddBlock(block, repo)

	if _, ok := errors.AsType[MoreOneCoinbaseError](err); !ok {
		t.Fatalf("ожидалась MoreOneCoinbaseError, получено: %v", err)
	}
}

func TestNewBlock_CoinbaseExisted(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := transaction.TransactionValidator{UtxoRepo: repo}
	bc := NewBlockchain(1, 1, &validator)
	block, err := bc.newBlock([]transaction.Transaction{}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

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
	repo := inmemory.NewRepository()
	validator := transaction.TransactionValidator{UtxoRepo: repo}
	bc := NewBlockchain(1, 1, &validator)

	genesisBlock := mineGenesisBlock(t, &bc, repo)
	block, err := bc.newBlock([]transaction.Transaction{}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	expectedHash := genesisBlock.Header.Hash()

	if block.Header.PrevHash != expectedHash {
		t.Fatalf("PrevHash не совпадает. Ожидалось: %v, получено: %v", expectedHash, block.Header.PrevHash)
	}
}

func TestNewBlock_ValidRootHash(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := transaction.TransactionValidator{UtxoRepo: repo}
	bc := NewBlockchain(1, 1, &validator)
	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{TxID: [32]byte{1}}}},
		Outputs: []coin.TxOutput{{Amount: 3, PublicKey: publicKey}},
	}
	coinbaseTx := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: bc.currentAward, PublicKey: publicKey}},
	}

	block, err := bc.newBlock([]transaction.Transaction{tx}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	root, err := merkle.BuildMerkleTree([]transaction.Transaction{tx, coinbaseTx})

	if err != nil {
		t.Fatal(err)
	}

	expectedHash := root.Hash

	if block.Header.RootHash != expectedHash {
		t.Fatalf("RootHash не совпадает. Ожидалось: %v, получено: %v", expectedHash, block.Header.RootHash)
	}
}

func TestAddBlock_CoinbaseExisted(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := transaction.TransactionValidator{UtxoRepo: repo}
	bc := NewBlockchain(1, 1, &validator)
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
	err := block.CalculateRootHash()

	if err != nil {
		t.Fatal(err)
	}

	block.Mine()

	err = bc.AddBlock(block, repo)

	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	utxos, err := repo.FindByOwner(publicKey)

	if err != nil {
		t.Fatalf("не удалось получить utxo: %v", err)
	}

	if len(utxos) == 0 {
		t.Fatalf("coinbase-транзакция не добавилась в repo")
	}

	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	assertUTXO(t, repo, utxo.UTXOID{TxID: hash}, bc.currentAward, publicKey)
}

func TestMineAndAddBlock_Success(t *testing.T) {
	privateKey, publicKey := testutil.GenerateKeyPair(t)
	repo := inmemory.NewRepository()
	validator := transaction.TransactionValidator{UtxoRepo: repo}
	_, otherPublicKey := testutil.GenerateKeyPair(t)
	bc := NewBlockchain(2, 5, &validator)
	mempool := mempool.NewMempool(&validator)

	genesisBlock, err := bc.newBlock([]transaction.Transaction{}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	genesisBlock.Mine()
	err = bc.AddBlock(genesisBlock, repo)

	if err != nil {
		t.Fatalf("ошибка при добавлении первого блока: %v", err)
	}

	genesisTxHash, err := genesisBlock.Transactions[0].Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{{
			ID: utxo.UTXOID{TxID: genesisTxHash},
		}},
		Outputs: []coin.TxOutput{{
			Amount: bc.currentAward, PublicKey: otherPublicKey,
		}},
	}
	txHash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, txHash)

	err = mempool.Add(tx, repo)

	if err != nil {
		t.Fatalf("ошибка при добавлении транзакции: %v", err)
	}

	_, err = bc.MineAndAddBlock(mempool, repo, 1, publicKey)

	if err != nil {
		t.Fatalf("ошибка при создании и добавлении блока: %v", err)
	}

	assertUTXO(t, repo, utxo.UTXOID{TxID: txHash}, bc.currentAward, otherPublicKey)
}

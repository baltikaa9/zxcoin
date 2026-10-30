package blockchain

import (
	"testing"

	"github.com/baltikaa9/zxcoin/adapters/persistence/inmemory"
	"github.com/baltikaa9/zxcoin/app/logger"
	"github.com/baltikaa9/zxcoin/app/mempool"
	apptransaction "github.com/baltikaa9/zxcoin/app/transaction"
	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/domain/blockchain"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
	"github.com/baltikaa9/zxcoin/core/testutil"
)

func newBlockchainService(t *testing.T, difficulty int, award int) *Service {
	t.Helper()
	bc := blockchain.NewBlockchain(uint64(difficulty), uint64(award))
	repo := inmemory.NewBlock()
	validator := apptransaction.NewValidator(repo)
	logger := logger.Nop()
	mp := mempool.NewMempool(validator, logger)

	return NewService(
		bc,
		repo,
		validator,
		mp,
		logger,
	)
}

func mineGenesisBlock(t *testing.T, bs *Service) block.Block {
	t.Helper()
	genesisBlock := block.Block{
		Header:       block.Header{},
		Transactions: []transaction.Transaction{},
	}

	if err := genesisBlock.CalculateRootHash(); err != nil {
		t.Fatal(err)
	}

	_ = genesisBlock.Mine(t.Context(), bs.blockchain.CurrentDifficulty())

	if err := bs.AddBlock(genesisBlock); err != nil {
		t.Fatalf("ошибка при добавлении генезис-блока: %v", err)
	}

	return genesisBlock
}

func assertUTXO(t *testing.T, repo apputxo.Repository, id utxo.ID, expectedAmount uint64, expectedOwner types.PublicKey) {
	t.Helper()
	u, existed, err := repo.FindByID(id)

	if err != nil {
		t.Fatalf("не удалось получить UTXO: %v", err)
	}

	if !existed {
		t.Fatalf("UTXO не найден: %v", id)
	}

	if u.Output.Amount != expectedAmount {
		t.Fatalf("неверная сумма UTXO. Ожидалось %v, получено %v", expectedAmount, u.Output.Amount)
	}

	if !u.Output.Owner.Equal(expectedOwner.PublicKey) {
		t.Fatalf("неверный владелец UTXO")
	}
}

func TestAddBlock_DoubleSpendInBlock(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, bs.utxoRepo)

	t1 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.ID{}}},
		Outputs: []coin.TxOutput{{Amount: amount, Owner: publicKey}},
	}
	t1Hash, err := t1.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if err := t1.Inputs[0].Sign(privateKey, t1Hash); err != nil {
		t.Fatal(err)
	}

	t2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.ID{}}},
		Outputs: []coin.TxOutput{{Amount: amount, Owner: publicKey}},
	}
	t2Hash, err := t2.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if err := t2.Inputs[0].Sign(privateKey, t2Hash); err != nil {
		t.Fatal(err)
	}

	block, err := bs.newBlock([]transaction.Transaction{t1, t2}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	_ = block.Mine(t.Context(), bs.blockchain.CurrentDifficulty())
	err = bs.AddBlock(block)

	testutil.CheckError[DoubleSpendError](t, err)
}

func TestAddBlock_InvalidNonce(t *testing.T) {
	bs := newBlockchainService(t, 2, 1)

	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := uint64(5)

	testutil.GenerateSingleUtxo(t, amount, publicKey, bs.utxoRepo)

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{ID: utxo.ID{}},
		},
		Outputs: []coin.TxOutput{
			{Amount: amount, Owner: publicKey},
		},
	}
	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Inputs[0].Sign(privateKey, hash); err != nil {
		t.Fatal(err)
	}

	b, err := bs.newBlock([]transaction.Transaction{tx}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	err = bs.AddBlock(b)

	testutil.CheckError[block.InvalidNonceError](t, err)
}

func TestAddBlock_InvalidPrevHash(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	mineGenesisBlock(t, bs)
	block := block.Block{
		Header: block.Header{
			PrevHash:  types.Hash{},
			RootHash:  types.Hash{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
	}

	if err := block.CalculateRootHash(); err != nil {
		t.Fatal(err)
	}

	_ = block.Mine(t.Context(), bs.blockchain.CurrentDifficulty())
	err := bs.AddBlock(block)

	testutil.CheckError[blockchain.InvalidPrevHashError](t, err)
}

func TestAddBlock_InvalidMerkleRootHash(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	genesisBlock := mineGenesisBlock(t, bs)
	b := block.Block{
		Header: block.Header{
			PrevHash:  genesisBlock.Header.Hash(),
			RootHash:  types.Hash{1},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
	}
	_ = b.Mine(t.Context(), bs.blockchain.CurrentDifficulty())
	err := bs.AddBlock(b)

	testutil.CheckError[block.InvalidMerkleRootError](t, err)
}

func TestAddBlock_MoreOneCoinbase(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	_, publicKey := testutil.GenerateKeyPair(t)
	block := block.Block{
		Header: block.Header{
			PrevHash:  types.Hash{},
			RootHash:  types.Hash{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{
			{Outputs: []coin.TxOutput{{Amount: bs.blockchain.CurrentAward(), Owner: publicKey}}},
			{Outputs: []coin.TxOutput{{Amount: bs.blockchain.CurrentAward(), Owner: publicKey}}},
		},
	}

	if err := block.CalculateRootHash(); err != nil {
		t.Fatal(err)
	}

	_ = block.Mine(t.Context(), bs.blockchain.CurrentDifficulty())
	err := bs.AddBlock(block)

	testutil.CheckError[MoreOneCoinbaseError](t, err)
}

func TestNewBlock_CoinbaseExisted(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	_, publicKey := testutil.GenerateKeyPair(t)
	block, err := bs.newBlock([]transaction.Transaction{}, publicKey)

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

	if !output.Owner.Equal(publicKey.PublicKey) {
		t.Fatalf("неверный получатель coinbase-транзакции")
	}

	if output.Amount != bs.blockchain.CurrentAward() {
		t.Fatalf("неверный размер coinbase-транзакции")
	}
}

func TestNewBlock_ValidPrevHash(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	_, publicKey := testutil.GenerateKeyPair(t)

	genesisBlock := mineGenesisBlock(t, bs)
	block, err := bs.newBlock([]transaction.Transaction{}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	expectedHash := genesisBlock.Header.Hash()

	if block.Header.PrevHash != expectedHash {
		t.Fatalf("PrevHash не совпадает. Ожидалось: %v, получено: %v", expectedHash, block.Header.PrevHash)
	}
}

func TestNewBlock_ValidRootHash(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	_, publicKey := testutil.GenerateKeyPair(t)
	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.ID{TxID: types.Hash{1}}}},
		Outputs: []coin.TxOutput{{Amount: 3, Owner: publicKey}},
	}

	b, err := bs.newBlock([]transaction.Transaction{tx}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	if err := b.ValidateRootHash(); err != nil {
		t.Fatalf("корень Merkle tree некорректен: %v", err)
	}

	blockWithoutCoinbase := b
	blockWithoutCoinbase.Transactions = b.Transactions[:len(b.Transactions)-1]

	if err := blockWithoutCoinbase.CalculateRootHash(); err != nil {
		t.Fatal(err)
	}

	if b.Header.RootHash == blockWithoutCoinbase.Header.RootHash {
		t.Fatal("RootHash не зависит от coinbase-транзакции")
	}
}

func TestAddBlock_CoinbaseExisted(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	_, publicKey := testutil.GenerateKeyPair(t)
	tx := transaction.Transaction{Outputs: []coin.TxOutput{{Amount: bs.blockchain.CurrentAward(), Owner: publicKey}}}
	block := block.Block{
		Header: block.Header{
			PrevHash:  types.Hash{},
			RootHash:  types.Hash{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{tx},
	}

	if err := block.CalculateRootHash(); err != nil {
		t.Fatal(err)
	}

	_ = block.Mine(t.Context(), bs.blockchain.CurrentDifficulty())

	if err := bs.AddBlock(block); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	utxos, err := bs.utxoRepo.FindByOwner(publicKey)

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

	assertUTXO(t, bs.utxoRepo, utxo.ID{TxID: hash}, bs.blockchain.CurrentAward(), publicKey)
}

func TestMineAndAddBlock_Success(t *testing.T) {
	bs := newBlockchainService(t, 2, 5)

	privateKey, publicKey := testutil.GenerateKeyPair(t)
	_, otherPublicKey := testutil.GenerateKeyPair(t)

	genesisBlock, err := bs.newBlock([]transaction.Transaction{}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	_ = genesisBlock.Mine(t.Context(), bs.blockchain.CurrentDifficulty())

	if err := bs.AddBlock(genesisBlock); err != nil {
		t.Fatalf("ошибка при добавлении первого блока: %v", err)
	}

	genesisTxHash, err := genesisBlock.Transactions[0].Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{{
			ID: utxo.ID{TxID: genesisTxHash},
		}},
		Outputs: []coin.TxOutput{{
			Amount: bs.blockchain.CurrentAward(), Owner: otherPublicKey,
		}},
	}
	txHash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Inputs[0].Sign(privateKey, txHash); err != nil {
		t.Fatal(err)
	}

	if err := bs.mempool.Add(tx); err != nil {
		t.Fatalf("ошибка при добавлении транзакции: %v", err)
	}

	_, err = bs.MineAndAddBlock(t.Context(), 1, publicKey)

	if err != nil {
		t.Fatalf("ошибка при создании и добавлении блока: %v", err)
	}

	assertUTXO(t, bs.utxoRepo, utxo.ID{TxID: txHash}, bs.blockchain.CurrentAward(), otherPublicKey)
}

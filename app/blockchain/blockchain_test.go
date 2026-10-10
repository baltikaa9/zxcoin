package blockchain

import (
	"errors"
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

func newBlockchainService(t *testing.T, difficulty int, award int) *BlockchainService {
	t.Helper()
	bc := blockchain.NewBlockchain(uint64(difficulty), uint64(award))
	repo := inmemory.NewRepository()
	validator := apptransaction.NewValidator(repo)
	logger := logger.Nop()
	mp := mempool.NewMempool(validator, logger)

	return NewBlockchainService(
		bc,
		repo,
		validator,
		mp,
		logger,
	)
}

func mineGenesisBlock(t *testing.T, bs *BlockchainService) block.Block {
	t.Helper()
	genesisBlock := block.Block{
		Header:       block.BlockHeader{},
		Transactions: []transaction.Transaction{},
	}

	if err := genesisBlock.CalculateRootHash(); err != nil {
		t.Fatal(err)
	}

	genesisBlock.Mine(bs.blockchain.CurrentDifficulty())

	if err := bs.AddBlock(genesisBlock); err != nil {
		t.Fatalf("ошибка при добавлении генезис-блока: %v", err)
	}

	return genesisBlock
}

func assertUTXO(t *testing.T, repo apputxo.Repository, id utxo.UTXOID, expectedAmount uint64, expectedOwner types.PublicKey) {
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

	if !u.Output.Owner.Equal(expectedOwner) {
		t.Fatalf("неверный владелец UTXO")
	}
}

func TestAddBlock_DoubleSpendInBlock(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, bs.utxoRepo)

	t1 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{}}},
		Outputs: []coin.TxOutput{{Amount: amount, Owner: publicKey}},
	}
	t1Hash, err := t1.Hash()

	if err != nil {
		t.Fatal(err)
	}

	t1.Inputs[0].Sign(privateKey, t1Hash)

	t2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{}}},
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

	block.Mine(bs.blockchain.CurrentDifficulty())
	err = bs.AddBlock(block)

	if _, ok := errors.AsType[DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_DoubleSpendInTransaction(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := uint64(5)
	testutil.GenerateSingleUtxo(t, amount, publicKey, bs.utxoRepo)

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{ID: utxo.UTXOID{}},
			{ID: utxo.UTXOID{}},
		},
		Outputs: []coin.TxOutput{
			{Amount: amount, Owner: publicKey},
		},
	}

	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	for i := range tx.Inputs {
		if err := tx.Inputs[i].Sign(privateKey, hash); err != nil {
			t.Fatal(err)
		}
	}

	block, err := bs.newBlock([]transaction.Transaction{tx}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	block.Mine(bs.blockchain.CurrentDifficulty())
	err = bs.AddBlock(block)

	if _, ok := errors.AsType[DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_InvalidNonce(t *testing.T) {
	bs := newBlockchainService(t, 2, 1)

	privateKey, publicKey := testutil.GenerateKeyPair(t)
	amount := uint64(5)

	testutil.GenerateSingleUtxo(t, amount, publicKey, bs.utxoRepo)

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{ID: utxo.UTXOID{}},
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

	if _, ok := errors.AsType[block.InvalidNonceError](err); !ok {
		t.Fatalf("ожидалась InvalidNonceError, получено: %v", err)
	}
}

func TestAddBlock_InvalidPrevHash(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	mineGenesisBlock(t, bs)
	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  types.Hash{},
			RootHash:  types.Hash{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
	}
	err := block.CalculateRootHash()

	if err != nil {
		t.Fatal(err)
	}

	block.Mine(bs.blockchain.CurrentDifficulty())
	err = bs.AddBlock(block)

	if _, ok := errors.AsType[blockchain.InvalidPrevHashError](err); !ok {
		t.Fatalf("ожидалась InvalidPrevHashError, получено: %v", err)
	}
}

func TestAddBlock_InvalidMerkleRootHash(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	genesisBlock := mineGenesisBlock(t, bs)
	b := block.Block{
		Header: block.BlockHeader{
			PrevHash:  genesisBlock.Header.Hash(),
			RootHash:  types.Hash{1},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
	}
	b.Mine(bs.blockchain.CurrentDifficulty())
	err := bs.AddBlock(b)

	if _, ok := errors.AsType[block.InvalidMerkleRootError](err); !ok {
		t.Fatalf("ожидалась InvalidMerkleRootError, получено: %v", err)
	}
}

func TestAddBlock_MoreOneCoinbase(t *testing.T) {
	bs := newBlockchainService(t, 1, 1)

	_, publicKey := testutil.GenerateKeyPair(t)
	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  types.Hash{},
			RootHash:  types.Hash{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{
			{Outputs: []coin.TxOutput{{Amount: bs.blockchain.CurrentAward(), Owner: publicKey}}},
			{Outputs: []coin.TxOutput{{Amount: bs.blockchain.CurrentAward(), Owner: publicKey}}},
		},
	}

	err := block.CalculateRootHash()

	if err != nil {
		t.Fatal(err)
	}

	block.Mine(bs.blockchain.CurrentDifficulty())
	err = bs.AddBlock(block)

	if _, ok := errors.AsType[MoreOneCoinbaseError](err); !ok {
		t.Fatalf("ожидалась MoreOneCoinbaseError, получено: %v", err)
	}
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

	if !output.Owner.Equal(publicKey) {
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
		Inputs:  []transaction.TxInput{{ID: utxo.UTXOID{TxID: types.Hash{1}}}},
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
		Header: block.BlockHeader{
			PrevHash:  types.Hash{},
			RootHash:  types.Hash{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{tx},
	}
	err := block.CalculateRootHash()

	if err != nil {
		t.Fatal(err)
	}

	block.Mine(bs.blockchain.CurrentDifficulty())

	err = bs.AddBlock(block)

	if err != nil {
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

	assertUTXO(t, bs.utxoRepo, utxo.UTXOID{TxID: hash}, bs.blockchain.CurrentAward(), publicKey)
}

func TestMineAndAddBlock_Success(t *testing.T) {
	bs := newBlockchainService(t, 2, 5)

	privateKey, publicKey := testutil.GenerateKeyPair(t)
	_, otherPublicKey := testutil.GenerateKeyPair(t)

	genesisBlock, err := bs.newBlock([]transaction.Transaction{}, publicKey)

	if err != nil {
		t.Fatal(err)
	}

	genesisBlock.Mine(bs.blockchain.CurrentDifficulty())
	err = bs.AddBlock(genesisBlock)

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
			Amount: bs.blockchain.CurrentAward(), Owner: otherPublicKey,
		}},
	}
	txHash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	tx.Inputs[0].Sign(privateKey, txHash)

	err = bs.mempool.Add(tx)

	if err != nil {
		t.Fatalf("ошибка при добавлении транзакции: %v", err)
	}

	_, err = bs.MineAndAddBlock(1, publicKey)

	if err != nil {
		t.Fatalf("ошибка при создании и добавлении блока: %v", err)
	}

	assertUTXO(t, bs.utxoRepo, utxo.UTXOID{TxID: txHash}, bs.blockchain.CurrentAward(), otherPublicKey)
}

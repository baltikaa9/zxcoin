package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"testing"
	"zxcoin/block"
	"zxcoin/coin"
	"zxcoin/mempool"
	"zxcoin/merkle"
	"zxcoin/transaction"
	"zxcoin/utxo"
)

func TestAddBlock_DoubleSpendInBlock(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: publicKey,
			},
			Reserved: false,
		},
	}

	t1 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: publicKey}},
	}
	t1.Inputs[0].Sign(privateKey, t1.Hash())

	t2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: publicKey}},
	}
	t2.Inputs[0].Sign(privateKey, t2.Hash())

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	block := bc.NewBlock([]transaction.Transaction{t1, t2}, publicKey)
	block.Mine()
	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_DoubleSpendInTransaction(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: publicKey,
			},
			Reserved: false,
		},
	}

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{TxID: [32]byte{}, OutIndex: 0},
			{TxID: [32]byte{}, OutIndex: 0},
		},
		Outputs: []coin.TxOutput{
			{Amount: 5, PublicKey: publicKey},
		},
	}

	hash := tx.Hash()

	for i := range tx.Inputs {
		tx.Inputs[i].Sign(privateKey, hash)
	}

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	block := bc.NewBlock([]transaction.Transaction{tx}, publicKey)
	block.Mine()
	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_InvalidNonce(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: publicKey,
			},
			Reserved: false,
		},
	}

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{TxID: [32]byte{}, OutIndex: 0},
		},
		Outputs: []coin.TxOutput{
			{Amount: 5, PublicKey: publicKey},
		},
	}
	tx.Inputs[0].Sign(privateKey, tx.Hash())

	bc := Blockchain{CurrentDifficulty: 2, CurrentAward: 1}
	block := bc.NewBlock([]transaction.Transaction{tx}, publicKey)

	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidNonceError](err); !ok {
		t.Fatalf("ожидалась InvalidNonceError, получено: %v", err)
	}
}

func TestAddBlock_InvalidPrevHash(t *testing.T) {
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := utxo.UTXODB{}

	genesisBlock := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
		Difficulty:   bc.CurrentDifficulty,
	}
	genesisBlock.CalculateRootHash()
	genesisBlock.Mine()

	err := bc.AddBlock(genesisBlock, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении первого блока: %v", err)
	}

	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
		Difficulty:   bc.CurrentDifficulty,
	}
	block.CalculateRootHash()
	block.Mine()

	err = bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidPrevHashError](err); !ok {
		t.Fatalf("ожидалась InvalidPrevHashError, получено: %v", err)
	}
}

func TestAddBlock_InvalidMerkleRootHash(t *testing.T) {
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := utxo.UTXODB{}

	genesisBlock := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
		Difficulty:   bc.CurrentDifficulty,
	}
	genesisBlock.CalculateRootHash()
	genesisBlock.Mine()

	err := bc.AddBlock(genesisBlock, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении первого блока: %v", err)
	}

	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  genesisBlock.Header.Hash(),
			RootHash:  [32]byte{1},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
		Difficulty:   bc.CurrentDifficulty,
	}
	block.Mine()

	err = bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidMerkleRootError](err); !ok {
		t.Fatalf("ожидалась InvalidMerkleRootError, получено: %v", err)
	}
}

func TestAddBlock_MoreOneCoinbase(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := utxo.UTXODB{}

	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{
			{Outputs: []coin.TxOutput{{Amount: bc.CurrentAward, PublicKey: publicKey}}},
			{Outputs: []coin.TxOutput{{Amount: bc.CurrentAward, PublicKey: publicKey}}},
		},
		Difficulty: bc.CurrentDifficulty,
	}
	block.CalculateRootHash()
	block.Mine()

	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*MoreOneCoinbaseError](err); !ok {
		t.Fatalf("ожидалась MoreOneCoinbaseError, получено: %v", err)
	}
}

func TestNewBlock_CoinbaseExisted(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	block := bc.NewBlock([]transaction.Transaction{}, publicKey)

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

	if output.Amount != bc.CurrentAward {
		t.Fatalf("неверный размер coinbase-транзакции")
	}
}

func TestNewBlock_ValidPrevHash(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}
	utxoDB := utxo.UTXODB{}

	genesisBlock := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
		Difficulty:   bc.CurrentDifficulty,
	}
	genesisBlock.CalculateRootHash()
	genesisBlock.Mine()

	err := bc.AddBlock(genesisBlock, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении первого блока: %v", err)
	}

	block := bc.NewBlock([]transaction.Transaction{}, publicKey)
	expectedHash := genesisBlock.Header.Hash()

	if block.Header.PrevHash != expectedHash {
		t.Fatalf("PrevHash не совпадает. Ожидалось: %v, получено: %v", expectedHash, block.Header.PrevHash)
	}
}

func TestNewBlock_ValidRootHash(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	tx := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{1}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 3, PublicKey: publicKey}},
	}
	coinbaseTx := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: bc.CurrentAward, PublicKey: publicKey}},
	}

	block := bc.NewBlock([]transaction.Transaction{tx}, publicKey)
	expectedHash := merkle.BuildMerkleTree([]transaction.Transaction{tx, coinbaseTx}).Hash

	if block.Header.RootHash != expectedHash {
		t.Fatalf("RootHash не совпадает. Ожидалось: %v, получено: %v", expectedHash, block.Header.RootHash)
	}
}

func TestAddBlock_CoinbaseExisted(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := utxo.UTXODB{}

	tx := transaction.Transaction{Outputs: []coin.TxOutput{{Amount: bc.CurrentAward, PublicKey: publicKey}}}

	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{tx},
		Difficulty:   bc.CurrentDifficulty,
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

	utxo, existed := utxoDB[utxo.UTXOKey{TxID: tx.Hash(), OutIndex: 0}]

	if !existed {
		t.Fatalf("некорректный ключ coinbase-транзакции в utxodb")
	}

	if utxo.Reserved {
		t.Fatalf("некорректная резервация coinbase-транзакции в utxodb")
	}

	if utxo.Output.Amount != bc.CurrentAward {
		t.Fatalf("некорректная сумма coinbase-транзакции в utxodb")
	}

	if !utxo.Output.PublicKey.Equal(publicKey) {
		t.Fatalf("некорректный получатель coinbase-транзакции в utxodb")
	}
}

func TestMineAndAddBlock_Success(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	otherPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	otherPublicKey := &otherPrivateKey.PublicKey

	bc := Blockchain{CurrentDifficulty: 2, CurrentAward: 5}
	utxoDB := utxo.UTXODB{}
	mempool := mempool.NewMempool()

	genesisBlock := bc.NewBlock([]transaction.Transaction{}, publicKey)
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
			{Amount: bc.CurrentAward, PublicKey: otherPublicKey},
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

	utxo, existed := utxoDB[utxo.UTXOKey{TxID: tx.Hash(), OutIndex: 0}]

	if !existed {
		t.Fatalf("некорректный ключ транзакции в utxodb")
	}

	if utxo.Reserved {
		t.Fatalf("некорректная резервация транзакции в utxodb")
	}

	if utxo.Output.Amount != bc.CurrentAward {
		t.Fatalf("некорректная сумма транзакции в utxodb")
	}

	if !utxo.Output.PublicKey.Equal(otherPublicKey) {
		t.Fatalf("некорректный получатель транзакции в utxodb")
	}
}

package blockchain

import (
	"errors"
	"testing"
	"zxcoin/block"
	"zxcoin/coin"
	"zxcoin/mempool"
	"zxcoin/transaction"
	"zxcoin/utxo"
	"zxcoin/wallet"
)

func TestAddBlock_DoubleSpendInBlock(t *testing.T) {
	myWallet := wallet.NewWallet()
	otherWallet := wallet.NewWallet()

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: myWallet.PublicKey,
			},
			Reserved: false,
		},
	}
	mempool := mempool.NewMempool()

	t1, err := myWallet.CreateTransaction(otherWallet.PublicKey, 5, utxoDB)

	if err != nil {
		t.Errorf("ошибка при создании честной транзакции: %v", err)
	}

	t2 := transaction.Transaction{
		Inputs:  []transaction.TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: myWallet.PublicKey}},
	}
	t2.Inputs[0].Sign(myWallet.PrivateKey, t2.Hash())

	err = mempool.Add(t1, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции в пул: %v", err)
	}

	err = mempool.Add(t2, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции в пул: %v", err)
	}

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	_, err = bc.MineAndAddBlock(mempool, utxoDB, 2, myWallet.PublicKey)

	if _, ok := errors.AsType[*DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_DoubleSpendInTransaction(t *testing.T) {
	myWallet := wallet.NewWallet()
	otherWallet := wallet.NewWallet()

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: myWallet.PublicKey,
			},
			Reserved: false,
		},
	}
	mempool := mempool.NewMempool()

	t1 := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{TxID: [32]byte{}, OutIndex: 0},
			{TxID: [32]byte{}, OutIndex: 0},
		},
		Outputs: []coin.TxOutput{
			{Amount: 5, PublicKey: otherWallet.PublicKey},
		},
	}

	hash := t1.Hash()

	for i := range t1.Inputs {
		t1.Inputs[i].Sign(myWallet.PrivateKey, hash)
	}

	err := mempool.Add(t1, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции: %v", err)
	}

	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	_, err = bc.MineAndAddBlock(mempool, utxoDB, 1, myWallet.PublicKey)

	if _, ok := errors.AsType[*DoubleSpendError](err); !ok {
		t.Fatalf("ожидалась DoubleSpendError, получено: %v", err)
	}
}

func TestAddBlock_InvalidNonce(t *testing.T) {
	myWallet := wallet.NewWallet()
	otherWallet := wallet.NewWallet()

	utxoDB := utxo.UTXODB{
		utxo.UTXOKey{
			TxID:     [32]byte{},
			OutIndex: 0,
		}: utxo.UTXOEntry{
			Output: coin.TxOutput{
				Amount:    5,
				PublicKey: myWallet.PublicKey,
			},
			Reserved: false,
		},
	}
	mempool := mempool.NewMempool()

	tx := transaction.Transaction{
		Inputs: []transaction.TxInput{
			{TxID: [32]byte{}, OutIndex: 0},
		},
		Outputs: []coin.TxOutput{
			{Amount: 5, PublicKey: otherWallet.PublicKey},
		},
	}
	tx.Inputs[0].Sign(myWallet.PrivateKey, tx.Hash())

	err := mempool.Add(tx, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции: %v", err)
	}

	bc := Blockchain{CurrentDifficulty: 2, CurrentAward: 1}
	block := bc.NewBlock(mempool.GetPending(1), myWallet.PublicKey)

	err = bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidNonceError](err); !ok {
		t.Fatalf("ожидалась InvalidNonceError, получено: %v", err)
	}
}

func TestAddBlock_InvalidPrevHash(t *testing.T) {
	myWallet := wallet.NewWallet()
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := utxo.UTXODB{}

	genesisBlock := bc.NewBlock([]transaction.Transaction{}, myWallet.PublicKey)
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
	block.Mine()

	err = bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidPrevHashError](err); !ok {
		t.Fatalf("ожидалась InvalidPrevHashError, получено: %v", err)
	}
}

func TestAddBlock_InvalidMerkleRootHash(t *testing.T) {
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := utxo.UTXODB{}

	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{1},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{},
		Difficulty:   bc.CurrentDifficulty,
	}
	block.Mine()

	err := bc.AddBlock(block, utxoDB)

	if _, ok := errors.AsType[*InvalidMerkleRootError](err); !ok {
		t.Fatalf("ожидалась InvalidMerkleRootError, получено: %v", err)
	}
}

func TestAddBlock_MoreOneCoinbase(t *testing.T) {
	myWallet := wallet.NewWallet()
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := utxo.UTXODB{}

	block := block.Block{
		Header: block.BlockHeader{
			PrevHash:  [32]byte{},
			RootHash:  [32]byte{},
			Timestamp: 0,
		},
		Transactions: []transaction.Transaction{
			{Outputs: []coin.TxOutput{{Amount: bc.CurrentAward, PublicKey: myWallet.PublicKey}}},
			{Outputs: []coin.TxOutput{{Amount: bc.CurrentAward, PublicKey: myWallet.PublicKey}}},
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

func TestAddBlock_CoinbaseExisted(t *testing.T) {
	myWallet := wallet.NewWallet()
	bc := Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	utxoDB := utxo.UTXODB{}

	block := bc.NewBlock([]transaction.Transaction{}, myWallet.PublicKey)

	if len(block.Transactions) < 1 {
		t.Fatalf("отсутствует coinbase-транзакция")
	}

	tx := block.Transactions[0]

	if len(tx.Inputs) > 0 || len(tx.Outputs) != 1 {
		t.Fatalf("некорректная coinbase-транзакция")
	}

	output := tx.Outputs[0]

	if output.PublicKey != myWallet.PublicKey {
		t.Fatalf("неверный получатель coinbase-транзакции")
	}

	if output.Amount != bc.CurrentAward {
		t.Fatalf("неверный размер coinbase-транзакции")
	}

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

	if !utxo.Output.PublicKey.Equal(myWallet.PublicKey) {
		t.Fatalf("некорректный получатель coinbase-транзакции в utxodb")
	}
}

func TestAddBlock_Success(t *testing.T) {
	myWallet := wallet.NewWallet()
	otherWallet := wallet.NewWallet()

	bc := Blockchain{CurrentDifficulty: 2, CurrentAward: 5}
	utxoDB := utxo.UTXODB{}
	mempool := mempool.NewMempool()

	genesisBlock := bc.NewBlock([]transaction.Transaction{}, myWallet.PublicKey)
	genesisBlock.Mine()

	err := bc.AddBlock(genesisBlock, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при добавлении первого блока: %v", err)
	}

	tx, err := myWallet.CreateTransaction(otherWallet.PublicKey, 5, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при создании транзакции: %v", err)
	}

	err = mempool.Add(tx, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при добавлении транзакции: %v", err)
	}

	_, err = bc.MineAndAddBlock(mempool, utxoDB, 1, myWallet.PublicKey)

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

	if !utxo.Output.PublicKey.Equal(otherWallet.PublicKey) {
		t.Fatalf("некорректный получатель транзакции в utxodb")
	}
}

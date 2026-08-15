package transaction

import (
	"errors"
	"testing"
	"zxcoin/blockchain"
	"zxcoin/coin"
	"zxcoin/mempool"
	"zxcoin/utxo"
	"zxcoin/wallet"
)

func TestValidate_UTXONotFound(t *testing.T) {
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

	err = mempool.Add(t1, utxoDB)

	if err != nil {
		t.Errorf("ошибка при добавлении транзакции: %v", err)
	}

	bc := blockchain.Blockchain{CurrentDifficulty: 1, CurrentAward: 1}

	_, err = bc.MineAndAddBlock(mempool, utxoDB, 1, myWallet.PublicKey)

	if err != nil {
		t.Fatalf("%v", err)
	}

	t2 := Transaction{[]TxInput{{TxID: [32]byte{}, OutIndex: 0}}, []coin.TxOutput{{Amount: 5, PublicKey: myWallet.PublicKey}}}
	t2.Inputs[0].Sign(myWallet.PrivateKey, t2.Hash())

	err = mempool.Add(t2, utxoDB)

	if _, ok := errors.AsType[*UTXONotFoundError](err); !ok {
		t.Fatalf("ожидалась UTXONotFoundError, получено: %v", err)
	}
}

func TestValidate_InvalidSignature(t *testing.T) {
	myWallet := wallet.NewWallet()
	otherWallet := wallet.NewWallet()
	attackerWallet := wallet.NewWallet()

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

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: otherWallet.PublicKey}},
	}
	tx.Inputs[0].Sign(attackerWallet.PrivateKey, tx.Hash())

	err := mempool.Add(tx, utxoDB)

	if _, ok := errors.AsType[*InvalidSignatureError](err); !ok {
		t.Fatalf("ожидалась InvalidSignatureError, получено: %v", err)
	}
}

func TestValidate_NotEnoughMoney(t *testing.T) {
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

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 10, PublicKey: otherWallet.PublicKey}},
	}
	tx.Inputs[0].Sign(myWallet.PrivateKey, tx.Hash())

	err := mempool.Add(tx, utxoDB)

	if _, ok := errors.AsType[*NotEnoughMoneyError](err); !ok {
		t.Fatalf("ожидалась NotEnoughMoney, получено: %v", err)
	}
}

func TestValidate_Success(t *testing.T) {
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

	tx := Transaction{
		Inputs:  []TxInput{{TxID: [32]byte{}, OutIndex: 0}},
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: otherWallet.PublicKey}},
	}
	tx.Inputs[0].Sign(myWallet.PrivateKey, tx.Hash())

	err := mempool.Add(tx, utxoDB)

	if err != nil {
		t.Fatalf("ошибка при валидации: %v", err)
	}
}

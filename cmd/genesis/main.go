package main

import (
	"context"
	"fmt"

	"github.com/baltikaa9/zxcoin/app/wallet"
	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
)

func main() {
	privateKey, err := wallet.LoadPrivateKey("wallet.key.test")
	if err != nil {
		panic(err)
	}

	tx, err := transaction.NewTransaction(
		[]transaction.TxInput{},
		[]coin.TxOutput{{
			Amount: 42,
			Owner:  types.PublicKey{PublicKey: &privateKey.PublicKey},
		}},
		privateKey,
	)
	if err != nil {
		panic(err)
	}

	b := block.Block{Transactions: []transaction.Transaction{tx}}

	if err := b.CalculateRootHash(); err != nil {
		panic(err)
	}

	if err := b.Mine(context.Background(), 0); err != nil {
		panic(err)
	}

	bHash := b.Header.Hash()

	txHash, err := tx.Hash()
	if err != nil {
		panic(err)
	}

	output := tx.Outputs[0]
	ownerKey, err := output.Owner.Bytes()
	if err != nil {
		panic(err)
	}

	fmt.Printf("UP\n==================================================\n\n")

	fmt.Printf(
		"INSERT INTO blocks (hash, height, prev_hash, root_hash, nonce, timestamp) VALUES (X'%x', %d, X'%x', X'%x', %d, %d);\n\n",
		bHash[:],
		0,
		b.Header.PrevHash[:],
		b.Header.RootHash[:],
		b.Header.Nonce,
		b.Header.Timestamp,
	)

	fmt.Printf(
		"INSERT INTO transactions (hash, block_hash, coinbase_height, position) VALUES (X'%x', X'%x', %d, %d);\n\n",
		txHash[:],
		bHash[:],
		0,
		0,
	)

	fmt.Printf(
		"INSERT INTO outputs (tx_id, out_index, amount, owner_key) VALUES (X'%x', %d, %d, X'%x');\n\n",
		txHash[:],
		0,
		output.Amount,
		ownerKey,
	)

	fmt.Printf("==================================================\nDOWN\n==================================================\n\n")

	fmt.Print(`CREATE TEMP TABLE _irreversible_error (id INT);

CREATE TEMP TRIGGER _trg_irreversible
BEFORE INSERT ON _irreversible_error
BEGIN
	SELECT RAISE(ABORT, 'Irreversible migration: cannot delete genesis block');
END;

INSERT INTO _irreversible_error VALUES (1);`,
	)

	fmt.Printf("\n\n==================================================")
}

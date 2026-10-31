package sqlite

import (
	"context"
	"database/sql"
	"errors"

	appblock "github.com/baltikaa9/zxcoin/app/block"
	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

const blockColumns = "height, prev_hash, root_hash, nonce, timestamp"

var _ appblock.Repository = (*Block)(nil)

type Block struct {
	db *sql.DB
}

func NewBlock(db *sql.DB) *Block {
	return &Block{db: db}
}

func (r *Block) FindByHash(hash types.Hash) (block.Block, bool, error) {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})

	if err != nil {
		return block.Block{}, false, err
	}

	defer tx.Rollback()

	b, _, exists, err := queryBlock(
		tx,
		`SELECT `+blockColumns+` FROM blocks WHERE hash = ?`,
		hash[:],
	)

	return b, exists, err
}

func (r *Block) Tip() (block.Block, uint64, bool, error) {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})

	if err != nil {
		return block.Block{}, 0, false, err
	}

	defer tx.Rollback()

	return queryBlock(
		tx,
		`SELECT `+blockColumns+` FROM blocks
		ORDER BY height DESC, rowid ASC
		LIMIT 1`,
	)
}

func (r *Block) FindByHeight(height uint64) ([]block.Block, error) {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	return queryBlocks(
		tx,
		`SELECT `+blockColumns+` FROM blocks WHERE height = ? ORDER BY hash`,
		height,
	)
}

func (r *Block) List(from uint64, limit uint64) ([]block.Block, error) {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	return queryBlocks(
		tx,
		`SELECT `+blockColumns+` FROM blocks WHERE height >= ? ORDER BY height, rowid LIMIT ?`,
		from,
		limit,
	)
}

func (r *Block) Save(b block.Block, height uint64) error {
	ownerBytes, err := u.Output.Owner.Bytes()

	if err != nil {
		return err
	}

	_, err = e.Exec(
		`INSERT INTO outputs (tx_id, out_index, amount, owner_key)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (tx_id, out_index) DO UPDATE SET
			amount = excluded.amount,
			owner_key = excluded.owner_key`,
		u.ID.TxID[:],
		u.ID.OutIndex,
		u.Output.Amount,
		ownerBytes,
	)

	return err
}

func loadBlockTransactions(tx *sql.Tx, hash types.Hash) ([]transaction.Transaction, error) {
	txs, index, err := loadTransactionStubs(tx, hash)

	if err != nil {
		return nil, err
	}

	if err := loadInputs(tx, hash, txs, index); err != nil {
		return nil, err
	}

	if err := loadOutputs(tx, hash, txs, index); err != nil {
		return nil, err
	}

	return txs, nil
}

func loadTransactionStubs(tx *sql.Tx, blockHash types.Hash) ([]transaction.Transaction, map[types.Hash]int, error) {
	rows, err := tx.Query(
		`SELECT hash, coinbase_height FROM transactions WHERE block_hash = ? ORDER BY position`,
		blockHash[:],
	)

	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	txs := make([]transaction.Transaction, 0)
	index := make(map[types.Hash]int)

	for rows.Next() {
		var hash []byte
		var coinbaseHeight uint64

		if err := rows.Scan(&hash, &coinbaseHeight); err != nil {
			return nil, nil, err
		}

		h, err := decodeHash(hash)

		if err != nil {
			return nil, nil, err
		}

		index[h] = len(txs)
		txs = append(txs, transaction.Transaction{CoinbaseHeight: coinbaseHeight})
	}

	return txs, index, rows.Err()
}

func loadInputs(tx *sql.Tx, blockHash types.Hash, txs []transaction.Transaction, index map[types.Hash]int) error {
	rows, err := tx.Query(
		`SELECT t.hash, i.tx_id, i.out_index, i.signature
		FROM transactions t
		JOIN inputs i ON i.tx_hash = t.hash
		WHERE t.block_hash = ?
		ORDER BY t.position, i.position`,
		blockHash[:],
	)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var txHash, prevTxID, sigBytes []byte
		var outIndex uint64

		if err := rows.Scan(&txHash, &prevTxID, &outIndex, &sigBytes); err != nil {
			return err
		}

		h, err := decodeHash(txHash)

		if err != nil {
			return err
		}

		prevID, err := decodeHash(prevTxID)

		if err != nil {
			return err
		}

		sig, err := transaction.ParseSignature(sigBytes)

		if err != nil {
			return err
		}

		i := index[h]
		txs[i].Inputs = append(txs[i].Inputs, transaction.TxInput{
			ID:        utxo.ID{TxID: prevID, OutIndex: outIndex},
			Signature: sig,
		})
	}

	return rows.Err()
}

func loadOutputs(tx *sql.Tx, blockHash types.Hash, txs []transaction.Transaction, index map[types.Hash]int) error {
	rows, err := tx.Query(
		`SELECT o.tx_id, o.amount, o.owner_key
		FROM transactions t
		JOIN outputs o ON o.tx_id = t.hash
		WHERE t.block_hash = ?
		ORDER BY t.position, o.out_index`,
		blockHash[:],
	)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var txID, ownerKey []byte
		var amount uint64

		if err := rows.Scan(&txID, &amount, &ownerKey); err != nil {
			return err
		}

		h, err := decodeHash(txID)

		if err != nil {
			return err
		}

		publicKey, err := decodePublicKey(ownerKey)

		if err != nil {
			return err
		}

		i := index[h]
		txs[i].Outputs = append(txs[i].Outputs, coin.TxOutput{
			Amount: amount,
			Owner:  publicKey,
		})
	}

	return rows.Err()
}

func queryBlock(tx *sql.Tx, query string, args ...any) (block.Block, uint64, bool, error) {
	var b block.Block
	var prevHash, rootHash []byte
	var height uint64

	if err := tx.QueryRow(query, args...).Scan(&height, &prevHash, &rootHash, &b.Header.Nonce, &b.Header.Timestamp); errors.Is(err, sql.ErrNoRows) {
		return block.Block{}, 0, false, nil
	} else if err != nil {
		return block.Block{}, 0, false, err
	}

	if err := decodeHeader(prevHash, rootHash, &b.Header); err != nil {
		return block.Block{}, 0, false, err
	}

	txs, err := loadBlockTransactions(tx, b.Header.Hash())

	if err != nil {
		return block.Block{}, 0, false, err
	}

	b.Transactions = txs

	return b, height, true, nil
}

func queryBlocks(tx *sql.Tx, query string, args ...any) ([]block.Block, error) {
	rows, err := tx.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	blocks := make([]block.Block, 0)

	for rows.Next() {
		var b block.Block
		var prevHash, rootHash []byte
		var height uint64

		if err := rows.Scan(&height, &prevHash, &rootHash, &b.Header.Nonce, &b.Header.Timestamp); err != nil {
			return nil, err
		}

		if err = decodeHeader(prevHash, rootHash, &b.Header); err != nil {
			return nil, err
		}

		blocks = append(blocks, b)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	for i, b := range blocks {
		txs, err := loadBlockTransactions(tx, b.Header.Hash())

		if err != nil {
			return nil, err
		}

		blocks[i].Transactions = txs
	}

	return blocks, nil
}

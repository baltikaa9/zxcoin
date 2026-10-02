package sqlite

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"database/sql"
	"fmt"

	"github.com/baltikaa9/zxcoin/coin"
	"github.com/baltikaa9/zxcoin/types"
	"github.com/baltikaa9/zxcoin/utxo"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByID(id utxo.UTXOID) (utxo.UTXO, bool, error) {
	return findByID(r.db, id)
}

func (r *Repository) FindByOwner(owner *ecdsa.PublicKey) ([]utxo.UTXO, error) {
	return findByOwner(r.db, owner)
}

func (r *Repository) Save(u utxo.UTXO) error {
	return save(r.db, u)
}

func (r *Repository) Delete(id utxo.UTXOID) error {
	return deleteByID(r.db, id)
}

func (r *Repository) Begin() (utxo.Transaction, error) {
	tx, err := r.db.Begin()

	if err != nil {
		return nil, err
	}

	return &transaction{
		db: tx,
	}, nil
}

type transaction struct {
	db *sql.Tx
}

func (t *transaction) FindByID(id utxo.UTXOID) (utxo.UTXO, bool, error) {
	return findByID(t.db, id)
}

func (t *transaction) FindByOwner(owner *ecdsa.PublicKey) ([]utxo.UTXO, error) {
	return findByOwner(t.db, owner)
}

func (t *transaction) Save(u utxo.UTXO) error {
	return save(t.db, u)
}

func (t *transaction) Delete(id utxo.UTXOID) error {
	return deleteByID(t.db, id)
}

func (t *transaction) Commit() error {
	return t.db.Commit()
}

func (t *transaction) Rollback() error {
	return t.db.Rollback()
}

type execer interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func findByID(e execer, id utxo.UTXOID) (utxo.UTXO, bool, error) {
	var amount uint64
	var ownerBytes []byte

	err := e.QueryRow(
		`SELECT amount, owner_key FROM utxos WHERE tx_id = ? AND out_index = ?`,
		id.TxID[:], id.OutIndex,
	).Scan(&amount, &ownerBytes)

	if err == sql.ErrNoRows {
		return utxo.UTXO{}, false, nil
	}

	if err != nil {
		return utxo.UTXO{}, false, err
	}

	owner, err := decodePublicKey(ownerBytes)

	if err != nil {
		return utxo.UTXO{}, false, err
	}

	return utxo.UTXO{
		ID:     id,
		Output: coin.TxOutput{Amount: amount, Owner: owner},
	}, true, nil
}

func findByOwner(e execer, owner *ecdsa.PublicKey) ([]utxo.UTXO, error) {
	ownerBytes, err := owner.Bytes()

	if err != nil {
		return nil, err
	}

	rows, err := e.Query(
		`SELECT tx_id, out_index, amount FROM utxos WHERE owner_key = ?`,
		ownerBytes,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]utxo.UTXO, 0)

	for rows.Next() {
		var txID []byte
		var outIndex, amount uint64

		if err := rows.Scan(&txID, &outIndex, &amount); err != nil {
			return nil, err
		}

		if len(txID) != len(types.Hash{}) {
			return nil, fmt.Errorf(
				"invalid tx_id length: got %d, want %d",
				len(txID),
				len(types.Hash{}),
			)
		}

		result = append(result, utxo.UTXO{
			ID:     utxo.UTXOID{TxID: types.Hash(txID), OutIndex: outIndex},
			Output: coin.TxOutput{Amount: amount, Owner: owner},
		})
	}

	return result, rows.Err()
}

func save(e execer, u utxo.UTXO) error {
	ownerBytes, err := u.Output.Owner.Bytes()

	if err != nil {
		return err
	}

	_, err = e.Exec(
		`INSERT INTO utxos (tx_id, out_index, amount, owner_key)
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

func deleteByID(e execer, id utxo.UTXOID) error {
	_, err := e.Exec(
		`DELETE FROM utxos WHERE tx_id = ? AND out_index = ?`,
		id.TxID[:], id.OutIndex,
	)

	return err
}

func decodePublicKey(b []byte) (*ecdsa.PublicKey, error) {
	return ecdsa.ParseUncompressedPublicKey(elliptic.P256(), b)
}

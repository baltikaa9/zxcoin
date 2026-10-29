package sqlite

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"database/sql"
	"errors"
	"fmt"

	apputxo "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

var _ apputxo.Repository = (*Repository)(nil)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByID(id utxo.ID) (utxo.UTXO, bool, error) {
	return findByID(r.db, id)
}

func (r *Repository) FindByOwner(owner types.PublicKey) ([]utxo.UTXO, error) {
	return findByOwner(r.db, owner)
}

func (r *Repository) Save(u utxo.UTXO) error {
	return save(r.db, u)
}

func (r *Repository) Delete(id utxo.ID) error {
	return deleteByID(r.db, id)
}

func (r *Repository) Begin() (apputxo.Transaction, error) {
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

func (t *transaction) FindByID(id utxo.ID) (utxo.UTXO, bool, error) {
	return findByID(t.db, id)
}

func (t *transaction) FindByOwner(owner types.PublicKey) ([]utxo.UTXO, error) {
	return findByOwner(t.db, owner)
}

func (t *transaction) Save(u utxo.UTXO) error {
	return save(t.db, u)
}

func (t *transaction) Delete(id utxo.ID) error {
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

func findByID(e execer, id utxo.ID) (utxo.UTXO, bool, error) {
	var amount uint64
	var ownerBytes []byte

	err := e.QueryRow(
		`SELECT amount, owner_key FROM outputs o WHERE tx_id = ? AND out_index = ? AND NOT EXISTS (
			SELECT 1 FROM inputs i WHERE i.tx_id = o.tx_id AND i.out_index = o.out_index
		)`,
		id.TxID[:], id.OutIndex,
	).Scan(&amount, &ownerBytes)

	if errors.Is(err, sql.ErrNoRows) {
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

func findByOwner(e execer, owner types.PublicKey) ([]utxo.UTXO, error) {
	ownerBytes, err := owner.Bytes()

	if err != nil {
		return nil, err
	}

	rows, err := e.Query(
		`SELECT tx_id, out_index, amount FROM outputs o WHERE owner_key = ? AND NOT EXISTS (
			SELECT 1 FROM inputs i WHERE i.tx_id = o.tx_id AND i.out_index = o.out_index
		)`,
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
			ID:     utxo.ID{TxID: types.Hash(txID), OutIndex: outIndex},
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

func deleteByID(e execer, id utxo.ID) error {
	return nil
}

func decodePublicKey(b []byte) (types.PublicKey, error) {
	key, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), b)

	if err != nil {
		return types.PublicKey{}, err
	}

	return types.PublicKey{PublicKey: key}, nil
}

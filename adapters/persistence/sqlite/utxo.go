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

var _ apputxo.Repository = (*UTXO)(nil)

type UTXO struct {
	db *sql.DB
}

func NewUTXO(db *sql.DB) *UTXO {
	return &UTXO{db: db}
}

func (r *UTXO) FindByID(id utxo.ID) (utxo.UTXO, bool, error) {
	var amount uint64
	var ownerBytes []byte

	err := r.db.QueryRow(
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

func (r *UTXO) FindByOwner(owner types.PublicKey) ([]utxo.UTXO, error) {
	ownerBytes, err := owner.Bytes()

	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(
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

func decodePublicKey(b []byte) (types.PublicKey, error) {
	key, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), b)

	if err != nil {
		return types.PublicKey{}, err
	}

	return types.PublicKey{PublicKey: key}, nil
}

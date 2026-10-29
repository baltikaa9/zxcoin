package sqlite

import (
	"database/sql"

	appblock "github.com/baltikaa9/zxcoin/app/block"
	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/domain/types"
)

var _ appblock.Repository = (*Block)(nil)

type Block struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Block {
	return &Block{db: db}
}

func (r *Block) FindByHash(hash types.Hash) (*block.Block, error) {}
func (r *Block) FindByHeight(height int) (*block.Block, error)    {}
func (r *Block) List(from int, limit int) ([]block.Block, error)  {}
func (r *Block) Tip() (*block.Block, int, error)                  {}

func (r *Block) Save(b block.Block, height int) error {
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

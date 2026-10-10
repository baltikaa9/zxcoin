// Package blockchain реализует основную логику консенсуса: проверку и добавление блоков, защиту от двойной траты и управление наградой за майнинг.
package blockchain

import (
	"fmt"
	"time"

	"github.com/baltikaa9/zxcoin/app/logger"
	"github.com/baltikaa9/zxcoin/app/mempool"
	apptransaction "github.com/baltikaa9/zxcoin/app/transaction"
	utxoapp "github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/domain/blockchain"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type BlockchainService struct {
	blockchain *blockchain.Blockchain
	utxoRepo   utxoapp.Repository
	validator  *apptransaction.TransactionValidator
	mempool    *mempool.Mempool
	logger     logger.Logger
}

func NewBlockchainService(
	blockchain *blockchain.Blockchain,
	utxoRepo utxoapp.Repository,
	validator *apptransaction.TransactionValidator,
	mempool *mempool.Mempool,
	logger logger.Logger,
) *BlockchainService {
	return &BlockchainService{
		blockchain: blockchain,
		utxoRepo:   utxoRepo,
		validator:  validator,
		mempool:    mempool,
		logger:     logger.Named("blockchain"),
	}
}

func (bs *BlockchainService) MineAndAddBlock(
	transactionLimit uint64,
	creator types.PublicKey,
) (block.Block, error) {
	bs.logger.Debug("mine and add block start")
	txs := bs.mempool.GetPending(transactionLimit)

	newBlock, err := bs.newBlock(txs, creator)

	if err != nil {
		return block.Block{}, fmt.Errorf("create block: %w", err)
	}

	bs.logger.Info("mining start", "block_hash", newBlock.Header.Hash().String())
	newBlock.Mine(bs.blockchain.CurrentDifficulty())
	bs.logger.Info("mining finish", "block_hash", newBlock.Header.Hash().String())

	if err := bs.AddBlock(newBlock); err != nil {
		return block.Block{}, fmt.Errorf("add block: %w", err)
	}

	for _, tx := range txs {
		hash, err := tx.Hash()

		if err != nil {
			return block.Block{}, err
		}

		bs.mempool.Remove(hash)
	}

	return newBlock, nil
}

func (bs *BlockchainService) AddBlock(block block.Block) error {
	if err := block.ValidateProofOfWork(bs.blockchain.CurrentDifficulty()); err != nil {
		return fmt.Errorf("validate PoW: %w", err)
	}

	if err := block.ValidateRootHash(); err != nil {
		return fmt.Errorf("validate root hash: %w", err)
	}

	if err := bs.verifyTransactions(block); err != nil {
		return fmt.Errorf("verify transactions: %w", err)
	}

	if err := bs.blockchain.VerifyPrevHash(block); err != nil {
		return fmt.Errorf("verify prev hash: %w", err)
	}

	if err := bs.applyBlock(block); err != nil {
		return fmt.Errorf("apply block: %w", err)
	}

	bs.blockchain.AppendBlock(block)
	bs.logger.Info("added block")

	return nil
}

func (bs *BlockchainService) newBlock(
	transactions []transaction.Transaction,
	creator types.PublicKey,
) (block.Block, error) {
	bs.logger.Debug("creating block", "creator", creator.Short())
	coinbaseTransaction := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: bs.blockchain.CurrentAward(), Owner: creator}},
	}
	bs.logger.Info("created coinbase transaction", "amount", bs.blockchain.CurrentAward(), "owner", creator.Short())

	newBlock := block.Block{
		Header: block.BlockHeader{
			PrevHash:  bs.blockchain.LastHash(),
			Nonce:     0,
			Timestamp: uint32(time.Now().Unix()),
		},
		Transactions: append(transactions, coinbaseTransaction),
	}

	err := newBlock.CalculateRootHash()

	if err != nil {
		return block.Block{}, err
	}

	bs.logger.Info("created block",
		"hash", newBlock.Header.Hash().String(),
		"prev_hash", newBlock.Header.PrevHash.String(),
		"root_hash", newBlock.Header.RootHash.String(),
		"nonce", newBlock.Header.Nonce,
		"timestamp", newBlock.Header.Timestamp,
	)

	return newBlock, nil
}

func (bs *BlockchainService) applyBlock(block block.Block) error {
	bs.logger.Debug("applying block", "hash", block.Header.Hash().String())
	tx, err := bs.utxoRepo.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, transaction := range block.Transactions {
		hash, err := transaction.Hash()

		if err != nil {
			return err
		}

		for _, input := range transaction.Inputs {
			if err := tx.Delete(input.ID); err != nil {
				return err
			}
		}

		for i, output := range transaction.Outputs {
			if err := tx.Save(utxo.UTXO{
				ID: utxo.UTXOID{TxID: hash, OutIndex: uint64(i)},
				Output: coin.TxOutput{
					Amount: output.Amount,
					Owner:  output.Owner,
				},
			}); err != nil {
				return err
			}
		}

		bs.logger.Info("applyed transaction", "tx_id", hash.String())
	}

	return tx.Commit()
}

func (bs *BlockchainService) verifyTransactions(block block.Block) error {
	spentInThisBlock := make(map[utxo.UTXOID]struct{})
	foundCoinbase := false

	for _, transaction := range block.Transactions {
		if len(transaction.Inputs) == 0 && len(transaction.Outputs) == 1 && transaction.Outputs[0].Amount == bs.blockchain.CurrentAward() {
			if foundCoinbase {
				return MoreOneCoinbaseError{}
			}

			foundCoinbase = true
			continue
		}

		if err := bs.validator.Validate(transaction); err != nil {
			return err
		}

		for _, input := range transaction.Inputs {
			if _, exists := spentInThisBlock[input.ID]; exists {
				return DoubleSpendError{input.ID.TxID, input.ID.OutIndex}
			}

			spentInThisBlock[input.ID] = struct{}{}
		}
	}

	return nil
}

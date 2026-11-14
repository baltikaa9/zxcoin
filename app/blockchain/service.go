// Package blockchain реализует основную логику консенсуса: проверку и добавление блоков, защиту от двойной траты и управление наградой за майнинг.
package blockchain

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	appblock "github.com/baltikaa9/zxcoin/app/block"
	"github.com/baltikaa9/zxcoin/app/logger"
	"github.com/baltikaa9/zxcoin/app/mempool"
	apptransaction "github.com/baltikaa9/zxcoin/app/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/domain/blockchain"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type miningJob struct {
	tipCtx     context.Context
	candidate  block.Block
	difficulty uint64
}

type Service struct {
	mu         sync.RWMutex
	tipCtx     context.Context
	tipCancel  context.CancelCauseFunc
	blockchain *blockchain.Blockchain
	blockRepo  appblock.Repository
	validator  *apptransaction.Validator
	mempool    *mempool.Mempool
	logger     logger.Logger
}

func NewService(
	blockchain *blockchain.Blockchain,
	blockRepo appblock.Repository,
	validator *apptransaction.Validator,
	mempool *mempool.Mempool,
	logger logger.Logger,
) *Service {
	ctx, cancel := context.WithCancelCause(context.Background())

	return &Service{
		tipCtx:     ctx,
		tipCancel:  cancel,
		blockchain: blockchain,
		blockRepo:  blockRepo,
		validator:  validator,
		mempool:    mempool,
		logger:     logger.Named("blockchain"),
	}
}

func (bs *Service) MineAndAddBlock(
	ctx context.Context,
	transactionLimit uint64,
	creator types.PublicKey,
) (block.Block, error) {
	bs.logger.Debug("mine and add block start")

	txs := bs.mempool.GetPending(transactionLimit)
	if len(txs) == 0 {
		return block.Block{}, errors.New("ожидающих транзакций не найдено")
	}

	job, err := bs.newJob(txs, creator)
	if err != nil {
		return block.Block{}, fmt.Errorf("create block: %w", err)
	}

	mineCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	stop := context.AfterFunc(job.tipCtx, func() {
		cancel(context.Cause(job.tipCtx))
	})
	defer stop()

	bs.logger.Info("mining start", "tx_count", len(txs), "difficulty", job.difficulty)
	start := time.Now()

	if err := job.candidate.Mine(mineCtx, job.difficulty); err != nil {
		return block.Block{}, err
	}

	bs.logger.Info(
		"mining finish",
		"nonce", job.candidate.Header.Nonce,
		"hash", job.candidate.Header.Hash().String(),
		"duration", time.Since(start),
	)

	if err := bs.AddBlock(job.candidate); err != nil {
		return block.Block{}, fmt.Errorf("add block: %w", err)
	}

	for _, tx := range txs {
		hash, err := tx.Hash()
		if err != nil {
			return block.Block{}, fmt.Errorf("hash transaction: %w", err)
		}

		bs.mempool.Remove(hash)
	}

	return job.candidate, nil
}

func (bs *Service) AddBlock(b block.Block) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	tip, tipHeight, exists, err := bs.blockRepo.Tip()
	if err != nil {
		return fmt.Errorf("get tip: %w", err)
	}
	if !exists {
		return EmptyBlockchainError{}
	}

	if b.Header.PrevHash != tip.Header.Hash() {
		return InvalidPrevHashError{}
	}

	if err := b.ValidateProofOfWork(bs.blockchain.CurrentDifficulty()); err != nil {
		return fmt.Errorf("validate PoW: %w", err)
	}

	if err := b.ValidateRootHash(); err != nil {
		return fmt.Errorf("validate root hash: %w", err)
	}

	if err := bs.verifyTransactions(b, tipHeight+1); err != nil {
		return fmt.Errorf("verify transactions: %w", err)
	}

	if err := bs.applyBlock(b, tipHeight+1); err != nil {
		return fmt.Errorf("apply block: %w", err)
	}

	bs.tipCancel(TipChangedError{})
	bs.tipCtx, bs.tipCancel = context.WithCancelCause(context.Background())
	bs.logger.Info("block added", "hash", b.Header.Hash().String())

	return nil
}

func (bs *Service) newBlock(
	transactions []transaction.Transaction,
	creator types.PublicKey,
	height uint64,
	prevHash types.Hash,
) (block.Block, error) {
	bs.logger.Debug("creating block", "creator", creator.Short())

	coinbaseTransaction := transaction.Transaction{
		CoinbaseHeight: height,
		Outputs:        []coin.TxOutput{{Amount: bs.blockchain.CurrentAward(), Owner: creator}},
	}
	bs.logger.Debug("coinbase transaction created", "amount", bs.blockchain.CurrentAward(), "owner", creator.Short())

	newBlock := block.Block{
		Header: block.Header{
			PrevHash:  prevHash,
			Nonce:     0,
			Timestamp: uint32(time.Now().Unix()),
		},
		Transactions: append(transactions, coinbaseTransaction),
	}

	if err := newBlock.CalculateRootHash(); err != nil {
		return block.Block{}, err
	}

	bs.logger.Debug("block created",
		"prev_hash", newBlock.Header.PrevHash.String(),
		"root_hash", newBlock.Header.RootHash.String(),
		"timestamp", newBlock.Header.Timestamp,
	)

	return newBlock, nil
}

func (bs *Service) applyBlock(b block.Block, height uint64) error {
	hashStr := b.Header.Hash().String()
	bs.logger.Debug("applying block", "hash", hashStr)

	if err := bs.blockRepo.Save(b, height); err != nil {
		return err
	}

	bs.logger.Debug("block applied", "hash", hashStr)

	return nil
}

func (bs *Service) verifyTransactions(b block.Block, coinbaseHeight uint64) error {
	spentInThisBlock := make(map[utxo.ID]struct{})
	foundCoinbase := false

	for _, transaction := range b.Transactions {
		if transaction.CoinbaseHeight != 0 {
			if len(transaction.Inputs) == 0 &&
				len(transaction.Outputs) == 1 &&
				transaction.Outputs[0].Amount == bs.blockchain.CurrentAward() {
				if foundCoinbase {
					return MoreOneCoinbaseError{}
				}

				if transaction.CoinbaseHeight != coinbaseHeight {
					return fmt.Errorf("coinbase_height error; got %v, want %v", transaction.CoinbaseHeight, coinbaseHeight)
				}

				foundCoinbase = true
				continue
			} else {
				return errors.New("некорректная coinbase-транзакция: ожидаются отсутствие входов, один выход и сумма, равная награде за блок")
			}
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

	if !foundCoinbase {
		return errors.New("отсутствует coinbase-транзакция")
	}

	return nil
}

func (bs *Service) newJob(
	txs []transaction.Transaction,
	creator types.PublicKey,
) (miningJob, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	tip, tipHeight, exists, err := bs.blockRepo.Tip()
	if err != nil {
		return miningJob{}, fmt.Errorf("get tip: %w", err)
	}
	if !exists {
		return miningJob{}, EmptyBlockchainError{}
	}
	newBlock, err := bs.newBlock(txs, creator, tipHeight+1, tip.Header.Hash())
	if err != nil {
		return miningJob{}, err
	}

	return miningJob{
		tipCtx:     bs.tipCtx,
		candidate:  newBlock,
		difficulty: bs.blockchain.CurrentDifficulty(),
	}, nil
}

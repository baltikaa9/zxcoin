package handlers

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/baltikaa9/zxcoin/adapters/http/handlers/dto"
	"github.com/baltikaa9/zxcoin/adapters/http/handlers/utils"
	"github.com/baltikaa9/zxcoin/app/blockchain"
	"github.com/baltikaa9/zxcoin/app/logger"
	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type Mining struct {
	blockchain *blockchain.Service
	logger     logger.Logger
}

func NewMining(blockchain *blockchain.Service, logger logger.Logger) *Mining {
	return &Mining{blockchain: blockchain, logger: logger.Named("http.mining")}
}

func (h *Mining) Mine(w http.ResponseWriter, r *http.Request) {
	var data dto.Mining

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&data); err != nil {
		utils.JSONError(w, fmt.Sprintf("parse error: %v", err), http.StatusBadRequest)
		h.logger.Warn("decode transaction", "err", err)
		return
	}

	if err := data.Validate(); err != nil {
		utils.JSONError(w, fmt.Sprintf("validation error: %v", err), http.StatusBadRequest)
		h.logger.Warn("validate mining data", "err", err)
		return
	}

	publicKeyBytes, err := hex.DecodeString(data.Creator)

	if err != nil {
		utils.JSONError(w, fmt.Sprintf("владелец должен быть в формате hex: %v", err), http.StatusBadRequest)
		h.logger.Warn("decode owner pubkey", "err", err)
		return
	}

	publicKey, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), publicKeyBytes)

	if err != nil {
		utils.JSONError(w, fmt.Sprintf("некорректный публичный ключ: %v", err), http.StatusBadRequest)
		h.logger.Warn("parse owner pubkey", "err", err)
		return
	}

	b, err := h.blockchain.MineAndAddBlock(r.Context(), data.TransactionLimit, types.PublicKey{PublicKey: publicKey})

	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			h.logger.Debug("mining request cancelled", "err", err)
			return
		}

		if _, ok := errors.AsType[blockchain.TipChangedError](err); ok {
			utils.JSONError(w, err.Error(), http.StatusConflict)
			h.logger.Warn("added block while mining", "err", err)
			return
		}

		utils.JSONError(w, fmt.Sprintf("mining error: %v", err), http.StatusBadRequest)
		h.logger.Warn("mining error", "err", err)
		return
	}

	w.WriteHeader(http.StatusCreated)

	response := dto.Block{
		Header: dto.Header{
			PrevHash:  b.Header.PrevHash.String(),
			Nonce:     b.Header.Nonce,
			RootHash:  b.Header.RootHash.String(),
			Timestamp: b.Header.Timestamp,
		},
		Transactions: make([]dto.Transaction, 0, len(b.Transactions)),
	}

	for i, tx := range b.Transactions {
		txDto, err := dto.TransactionToDto(tx)

		if err != nil {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Блок успешно добавлен, но не удалось сформировать ответ",
			})
			h.logger.Error("transaction to dto", "tx_index", i, "err", err)
			return
		}

		response.Transactions = append(response.Transactions, txDto)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("encode response dto", "err", err)
	}
}

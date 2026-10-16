package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/baltikaa9/zxcoin/adapters/http/handlers/dto"
	"github.com/baltikaa9/zxcoin/adapters/http/handlers/utils"
	"github.com/baltikaa9/zxcoin/app/logger"
	"github.com/baltikaa9/zxcoin/app/mempool"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type Mempool struct {
	mempool *mempool.Mempool
	logger  logger.Logger
}

func NewMempool(mempool *mempool.Mempool, logger logger.Logger) *Mempool {
	return &Mempool{mempool: mempool, logger: logger.Named("http.mempool")}
}

func (h *Mempool) AddTransaction(w http.ResponseWriter, r *http.Request) {
	var data dto.Transaction

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&data); err != nil {
		utils.JSONError(w, fmt.Sprintf("parse error: %v", err), http.StatusBadRequest)
		h.logger.Warn("decode transaction", "err", err)
		return
	}

	if err := data.Validate(); err != nil {
		utils.JSONError(w, fmt.Sprintf("validation error: %v", err), http.StatusBadRequest)
		h.logger.Warn("validate transaction", "err", err)
		return
	}

	var inputs []transaction.TxInput
	var outputs []coin.TxOutput

	for _, input := range data.Inputs {
		txIDBytes, _ := hex.DecodeString(input.TxID)
		signatureBytes, _ := hex.DecodeString(input.Signature)
		signature, err := transaction.ParseSignature(signatureBytes)

		if err != nil {
			utils.JSONError(w, fmt.Sprintf("parse signature error: %v", err), http.StatusBadRequest)
			h.logger.Warn("parse signature", "err", err)
			return
		}

		inputs = append(inputs, transaction.TxInput{
			ID: utxo.ID{
				TxID:     types.Hash(txIDBytes),
				OutIndex: input.OutIndex,
			},
			Signature: signature,
		})
	}

	for _, output := range data.Outputs {
		publicKeyBytes, err := hex.DecodeString(output.PublicKey)

		if err != nil {
			utils.JSONError(w, fmt.Sprintf("владелец должен быть в формате hex: %v", err), http.StatusBadRequest)
			h.logger.Warn("decode output pubkey", "err", err)
			return
		}

		publicKey, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), publicKeyBytes)

		if err != nil {
			utils.JSONError(w, fmt.Sprintf("некорректный публичный ключ: %v", err), http.StatusBadRequest)
			h.logger.Warn("parse output pubkey", "err", err)
			return
		}

		outputs = append(outputs, coin.TxOutput{
			Amount: output.Amount,
			Owner:  types.PublicKey{PublicKey: publicKey},
		})
	}

	tx := transaction.Transaction{Inputs: inputs, Outputs: outputs}

	if err := h.mempool.Add(tx); err != nil {
		utils.JSONError(w, fmt.Sprintf("не удалось добавить транзакцию в пул: %v", err), http.StatusUnprocessableEntity)
		h.logger.Warn("add transaction", "err", err)
		return
	}

	response, err := dto.TransactionToDto(tx)

	if err != nil {
		utils.JSONError(w, "внутренняя ошибка", http.StatusInternalServerError)
		h.logger.Error("transaction to dto", "err", err)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.JSONError(w, "внутренняя ошибка", http.StatusInternalServerError)
		h.logger.Error("encode response dto", "err", err)
	}
}

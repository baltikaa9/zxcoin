package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/baltikaa9/zxcoin/app/mempool"
	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/domain/utxo"
)

type InputDto struct {
	TxID      string `json:"txID"`
	OutIndex  uint64 `json:"outIndex"`
	Signature string `json:"signature"`
}

type OutputDto struct {
	Amount    uint64 `json:"amount"`
	PublicKey string `json:"publicKey"`
}

type TransactionDto struct {
	Inputs  []InputDto  `json:"inputs"`
	Outputs []OutputDto `json:"outputs"`
}

type MempoolHandler struct {
	mempool *mempool.Mempool
}

func NewMempoolHandler(
	mempool *mempool.Mempool,
) *MempoolHandler {
	return &MempoolHandler{
		mempool: mempool,
	}
}

func (h *MempoolHandler) AddTransaction(w http.ResponseWriter, r *http.Request) {
	var data TransactionDto

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("parse error: %v", err), http.StatusBadRequest)
		return
	}

	var inputs []transaction.TxInput
	var outputs []coin.TxOutput

	for _, input := range data.Inputs {
		txIDBytes, err := hex.DecodeString(input.TxID)

		if err != nil {
			http.Error(w, fmt.Sprintf("decode txID error: %v", err), http.StatusBadRequest)
			return
		}

		if len(txIDBytes) != 32 {
			http.Error(w, fmt.Sprintf("invalid txID len: %d", len(txIDBytes)), http.StatusBadRequest)
			return
		}

		signatureBytes, err := hex.DecodeString(input.Signature)

		if err != nil {
			http.Error(w, fmt.Sprintf("decode signature error: %v", err), http.StatusBadRequest)
			return
		}

		signature, err := transaction.ParseSignature(signatureBytes)

		if err != nil {
			http.Error(w, fmt.Sprintf("parse signature error: %v", err), http.StatusBadRequest)
			return
		}

		inputs = append(inputs, transaction.TxInput{
			ID: utxo.UTXOID{
				TxID:     types.Hash(txIDBytes),
				OutIndex: input.OutIndex,
			},
			Signature: signature,
		})
	}

	for _, output := range data.Outputs {
		publicKeyBytes, err := hex.DecodeString(output.PublicKey)

		if err != nil {
			http.Error(w, fmt.Sprintf("не удалось декодировать публичный ключ: %v", err), http.StatusBadRequest)
			return
		}

		publicKey, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), publicKeyBytes)

		if err != nil {
			http.Error(w, fmt.Sprintf("не удалось спарсить публичный ключ: %v", err), http.StatusBadRequest)
			return
		}

		outputs = append(outputs, coin.TxOutput{
			Amount: output.Amount,
			Owner:  publicKey,
		})
	}

	tx := transaction.Transaction{Inputs: inputs, Outputs: outputs}

	if err := h.mempool.Add(tx); err != nil {
		http.Error(w, fmt.Sprintf("не удалось добавить транзакцию в пул: %v", err), http.StatusUnprocessableEntity)
		return
	}

	response, err := transactionToDto(tx)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("не удалось подготовить ответ: %v", err),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func transactionToDto(
	tx transaction.Transaction,
) (TransactionDto, error) {
	dto := TransactionDto{
		Inputs:  make([]InputDto, 0, len(tx.Inputs)),
		Outputs: make([]OutputDto, 0, len(tx.Outputs)),
	}

	for _, input := range tx.Inputs {
		dto.Inputs = append(dto.Inputs, InputDto{
			TxID:      input.ID.TxID.String(),
			OutIndex:  input.ID.OutIndex,
			Signature: hex.EncodeToString(input.Signature.Marshal()),
		})
	}

	for _, output := range tx.Outputs {
		publicKey, err := output.Owner.Bytes()
		if err != nil {
			return TransactionDto{}, err
		}

		dto.Outputs = append(dto.Outputs, OutputDto{
			Amount:    output.Amount,
			PublicKey: hex.EncodeToString(publicKey),
		})
	}

	return dto, nil
}

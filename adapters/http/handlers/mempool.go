package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

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

func (d TransactionDto) Validate() error {
	if len(d.Inputs) == 0 {
		return errors.New("не заполнено поле inputs")
	}

	if len(d.Outputs) == 0 {
		return errors.New("не заполнено поле outputs")
	}

	for i, input := range d.Inputs {
		if err := validateHexField(
			fmt.Sprintf("inputs[%d].txID", i),
			input.TxID,
			32,
		); err != nil {
			return err
		}

		if err := validateHexField(
			fmt.Sprintf("inputs[%d].signature", i),
			input.Signature,
			64,
		); err != nil {
			return err
		}
	}

	for i, output := range d.Outputs {
		if output.Amount == 0 {
			return fmt.Errorf("outputs[%d].amount must be greater than zero", i)
		}

		if err := validateHexField(
			fmt.Sprintf("outputs[%d].publicKey", i),
			output.PublicKey,
			65,
		); err != nil {
			return err
		}

		if output.PublicKey[:2] != "04" {
			return fmt.Errorf("outputs[%d].publicKey must be uncompressed", i)
		}
	}

	return nil
}

func validateHexField(field string, value string, expectedBytes int) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}

	decoded, err := hex.DecodeString(value)

	if err == nil {
		return fmt.Errorf("%s must be valid hex", field)
	}

	if len(decoded) != expectedBytes {
		return fmt.Errorf(
			"%s must contain %d bytes, got %d",
			field,
			expectedBytes,
			len(decoded),
		)
	}

	return nil
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

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("parse error: %v", err), http.StatusBadRequest)
		return
	}

	if err := data.Validate(); err != nil {
		http.Error(
			w,
			fmt.Sprintf("validation error: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	var inputs []transaction.TxInput
	var outputs []coin.TxOutput

	for _, input := range data.Inputs {
		txIDBytes, _ := hex.DecodeString(input.TxID)
		signatureBytes, _ := hex.DecodeString(input.Signature)
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
		publicKeyBytes, _ := hex.DecodeString(output.PublicKey)
		publicKey, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), publicKeyBytes)

		if err != nil {
			http.Error(w, fmt.Sprintf("не удалось спарсить публичный ключ: %v", err), http.StatusBadRequest)
			return
		}

		outputs = append(outputs, coin.TxOutput{
			Amount: output.Amount,
			Owner:  types.PublicKey{PublicKey: publicKey},
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

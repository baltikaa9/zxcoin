package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/baltikaa9/zxcoin/app/utxo"
)

type UTXODTO struct {
	TxID     string `json:"txID"`
	OutIndex uint64 `json:"outIndex"`
	Amount   uint64 `json:"amount"`
}

type UTXOResponseDTO struct {
	UTXOs []UTXODTO `json:"utxos"`
	Sum   uint64    `json:"sum"`
}

type UTXOHandler struct {
	repo utxo.Repository
}

func NewUTXOHandler(repo utxo.Repository) *UTXOHandler {
	return &UTXOHandler{repo: repo}
}

func (h *UTXOHandler) GetUTXOByOwner(w http.ResponseWriter, r *http.Request) {
	publicKeyBytes, err := hex.DecodeString(r.PathValue("owner"))

	if err != nil {
		http.Error(w, fmt.Sprintf("владелец должен быть в формате hex: %v", err), http.StatusBadRequest)
		return
	}

	publicKey, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), publicKeyBytes)

	if err != nil {
		http.Error(w, fmt.Sprintf("некорректный публичный ключ: %v", err), http.StatusBadRequest)
		return
	}

	utxos, err := h.repo.FindByOwner(publicKey)

	if err != nil {
		http.Error(w, "не удалось найти UTXO", http.StatusInternalServerError)
		return
	}

	utxosDTO := make([]UTXODTO, 0, len(utxos))
	var sum uint64

	for _, u := range utxos {
		sum += u.Output.Amount

		utxosDTO = append(utxosDTO, UTXODTO{
			TxID:     u.ID.TxID.String(),
			OutIndex: u.ID.OutIndex,
			Amount:   u.Output.Amount,
		})
	}

	if err := json.NewEncoder(w).Encode(UTXOResponseDTO{UTXOs: utxosDTO, Sum: sum}); err != nil {
		return
	}
}

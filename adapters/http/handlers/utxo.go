package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/baltikaa9/zxcoin/app/utxo"
)

type UTXODTO struct {
	TxID     string `json:"txID"`
	OutIndex uint64 `json:"outIndex"`
	Amount   uint64 `json:"amount"`
}

type ResponseDTO struct {
	UTXOS []UTXODTO `json:"utxos"`
	Sum   uint64    `json:"sum"`
}

type UTXOHandler struct {
	repo utxo.Repository
}

func NewUTXOHandler(repo utxo.Repository) *UTXOHandler {
	return &UTXOHandler{repo: repo}
}

func (h *UTXOHandler) GetUTXOByOwner(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")

	if strings.TrimSpace(owner) == "" {
		http.Error(w, "не указан публичный ключ владельца", http.StatusBadRequest)
		return
	}

	publicKeyBytes, err := hex.DecodeString(owner)

	if err != nil {
		http.Error(w, "владелец должен быть в формате hex", http.StatusBadRequest)
		return
	}

	publicKey, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), publicKeyBytes)

	if err != nil {
		http.Error(w, "некорректный публичный ключ", http.StatusBadRequest)
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

	if err := json.NewEncoder(w).Encode(ResponseDTO{UTXOS: utxosDTO, Sum: sum}); err != nil {
		return
	}
}

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
	"github.com/baltikaa9/zxcoin/app/utxo"
	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type UTXO struct {
	repo   utxo.Repository
	logger logger.Logger
}

func NewUTXO(repo utxo.Repository, logger logger.Logger) *UTXO {
	return &UTXO{repo: repo, logger: logger.Named("http.utxo")}
}

func (h *UTXO) GetUTXOByOwner(w http.ResponseWriter, r *http.Request) {
	publicKeyBytes, err := hex.DecodeString(r.PathValue("owner"))

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

	utxos, err := h.repo.FindByOwner(types.PublicKey{PublicKey: publicKey})

	if err != nil {
		utils.JSONError(w, "внутренняя ошибка", http.StatusInternalServerError)
		h.logger.Error("find utxo by owner", "err", err)
		return
	}

	utxosDTO := make([]dto.UTXO, 0, len(utxos))
	var sum uint64

	for _, u := range utxos {
		sum += u.Output.Amount

		utxosDTO = append(utxosDTO, dto.UTXO{
			TxID:     u.ID.TxID.String(),
			OutIndex: u.ID.OutIndex,
			Amount:   u.Output.Amount,
		})
	}

	if err := json.NewEncoder(w).Encode(dto.UTXOResponse{UTXOs: utxosDTO, Sum: sum}); err != nil {
		h.logger.Error("encode response dto", "err", err)
	}
}

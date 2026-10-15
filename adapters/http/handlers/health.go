package handlers

import (
	"net/http"

	"github.com/baltikaa9/zxcoin/adapters/http/handlers/utils"
)

func Health(w http.ResponseWriter, _ *http.Request) {
	if _, err := w.Write([]byte("ok")); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
	}
}

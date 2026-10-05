package handlers

import "net/http"

func Health(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("ok")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

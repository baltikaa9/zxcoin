package utils

import (
	"encoding/json"
	"net/http"
)

type httpError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func JSONError(w http.ResponseWriter, msg string, code int) {
	h := w.Header()

	// Delete the Content-Length header, which might be for some other content.
	// Assuming the error string fits in the writer's buffer, we'll figure
	// out the correct Content-Length for it later.
	//
	// We don't delete Content-Encoding, because some middleware sets
	// Content-Encoding: gzip and wraps the ResponseWriter to compress on-the-fly.
	// See https://go.dev/issue/66343.
	h.Del("Content-Length")

	// There might be content type already set, but we reset it to
	// application/json for the error message.
	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(httpError{msg, code}); err != nil {
		http.Error(w, msg, code)
	}
}

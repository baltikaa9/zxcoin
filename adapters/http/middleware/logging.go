package middleware

import (
	"net/http"
	"time"

	"github.com/baltikaa9/zxcoin/app/logger"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func Logging(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			args := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration", time.Since(start),
			}

			switch {
			case sw.status >= 500:
				log.Error("http request", args...)
			case sw.status >= 400:
				log.Warn("http request", args...)
			default:
				log.Info("http request", args...)
			}
		})
	}
}

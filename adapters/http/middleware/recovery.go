package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/baltikaa9/zxcoin/app/logger"
)

func Recovery(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()

				if rec == nil {
					return
				}

				if rec == http.ErrAbortHandler {
					panic(rec)
				}

				log.Error(
					"panic recovered",
					"panic", rec,
					"stack", string(debug.Stack()),
				)

				http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
			}()

			next.ServeHTTP(w, r)
		})
	}
}

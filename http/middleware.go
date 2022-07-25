package http

import (
	"net/http"

	"go.strv.io/net/logger"
)

func RecoverMiddleware(l logger.ServerLogger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if re := recover(); re != nil {
					l.With(logger.Any("err", re)).Error("panic recover", nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

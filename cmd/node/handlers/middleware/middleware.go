// Package middleware
package middleware

import (
	"log/slog"
	"net/http"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("request", "method", r.Method, "path", r.RequestURI)

		next.ServeHTTP(w, r)
	})
}

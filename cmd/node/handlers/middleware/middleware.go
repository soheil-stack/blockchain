// Package middleware
package middleware

import (
	"net/http"
)

func LoggerMiddleware(next http.Handler, logger any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log here

		next.ServeHTTP(w, r)
	})
}

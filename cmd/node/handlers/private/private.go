// Package private
package private

import (
	"net/http"

	"github.com/soheil-stack/blockchain/cmd/node/handlers/middleware"
	"github.com/soheil-stack/blockchain/internal/state"
)

func NewServer(s *state.State) http.Handler {
	mux := http.NewServeMux()

	var handler http.Handler = mux
	handler = middleware.LoggerMiddleware(handler, nil)

	return handler
}

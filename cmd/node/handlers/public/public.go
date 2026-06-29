// Package public
package public

import (
	"net/http"

	"github.com/soheil-stack/blockchain/cmd/node/handlers/middleware"
	"github.com/soheil-stack/blockchain/internal/nameservice"
	"github.com/soheil-stack/blockchain/internal/state"
)

func NewServer(s *state.State, ns *nameservice.NameService) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /genesis", GetGenesis(s))
	mux.Handle("GET /accounts", GetAccounts(s, ns))
	mux.Handle("GET /accounts/{address}", GetAccount(s, ns))
	mux.Handle("GET /mempool/transactions", GetMempoolTransactions(s, ns))
	mux.Handle("POST /transactions", PostTransaction(s))

	var handler http.Handler = mux
	handler = middleware.LoggerMiddleware(handler)

	return handler
}

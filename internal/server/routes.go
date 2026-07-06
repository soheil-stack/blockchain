// Package server
package server

import (
	"net/http"

	"github.com/soheil-stack/blockchain/internal/nameservice"
	"github.com/soheil-stack/blockchain/internal/state"
)

func New(s *state.State, ns *nameservice.NameService) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /genesis", GetGenesis(s))
	mux.Handle("GET /accounts", GetAccounts(s, ns))
	mux.Handle("GET /accounts/{address}", GetAccount(s, ns))
	mux.Handle("GET /mempool/transactions", GetMempoolTransactions(s, ns))
	mux.Handle("POST /transactions", PostTransaction(s))
	mux.Handle("POST /node/peers", PostPeer(s))
	mux.Handle("GET /node/status", GetStatus(s))
	mux.Handle("GET /blocks", GetBlocks(s))
	mux.Handle("POST /blocks", PostBlock(s))

	var handler http.Handler = mux
	handler = loggerMiddleware(handler)

	return handler
}

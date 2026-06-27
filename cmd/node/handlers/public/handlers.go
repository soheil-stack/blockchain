package public

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/cmd/node/handlers"
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/soheil-stack/blockchain/internal/state"
)

func GetGenesis(s *state.State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		genesis := s.Genesis()
		if err := handlers.Encode(w, r, genesis); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func GetAccounts(s *state.State, ns *core.NameService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accounts := s.Accounts()

		response := make([]AccountResponse, 0, len(accounts))
		for _, account := range accounts {
			response = append(response, toAccountResponse(ns, account))
		}

		if err := handlers.Encode(w, r, response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func GetAccount(s *state.State, ns *core.NameService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.PathValue("address"))
		account, ok := s.Account(address)
		if !ok {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}

		response := toAccountResponse(ns, account)

		if err := handlers.Encode(w, r, response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func GetMempoolTransactions(s *state.State, ns *core.NameService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		sender := common.HexToAddress(q.Get("sender"))
		receiver := common.HexToAddress(q.Get("receiver"))
		zeroAddress := common.Address{}

		mempool := s.Mempool()

		response := make([]TransactionResponse, 0)
		for _, tx := range mempool {

			if sender != zeroAddress && tx.From != sender {
				continue
			}

			if receiver != zeroAddress && tx.To != receiver {
				continue
			}

			transaction, err := toTransactionResponse(ns, tx)
			if err != nil {
				http.Error(w, "failed to encode response", http.StatusInternalServerError)
				return
			}
			response = append(response, transaction)
		}

		if err := handlers.Encode(w, r, response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func PostTransaction(s *state.State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx, err := handlers.Decode[core.Transaction](r)
		if err != nil {
			http.Error(w, "failed to decode request payload", http.StatusBadRequest)
			return
		}

		err = s.UpsertTransaction(tx)
		if err != nil {
			http.Error(w, "failed to post transaction", http.StatusInternalServerError)
		}

		response := struct {
			Status string `json:"status"`
		}{
			"transaction added to mempool",
		}

		if err := handlers.Encode(w, r, response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

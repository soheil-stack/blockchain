package server

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/soheil-stack/blockchain/internal/nameservice"
	"github.com/soheil-stack/blockchain/internal/peer"
	"github.com/soheil-stack/blockchain/internal/state"
)

func GetGenesis(s *state.State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		genesis := s.Genesis()
		if err := encode(w, r, genesis); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func GetAccounts(s *state.State, ns *nameservice.NameService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accounts := s.Accounts()

		response := make([]AccountResponse, 0, len(accounts))
		for _, account := range accounts {
			response = append(response, toAccountResponse(ns, account))
		}

		if err := encode(w, r, response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func GetAccount(s *state.State, ns *nameservice.NameService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.PathValue("address"))
		account, ok := s.Account(address)
		if !ok {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}

		response := toAccountResponse(ns, account)

		if err := encode(w, r, response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func GetMempoolTransactions(s *state.State, ns *nameservice.NameService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		from := common.HexToAddress(q.Get("from"))
		to := common.HexToAddress(q.Get("to"))
		zeroAddress := common.Address{}

		mempool := s.MempoolPickBest()

		response := make([]TransactionResponse, 0)
		for _, tx := range mempool {

			if from != zeroAddress && tx.From != from {
				continue
			}

			if to != zeroAddress && tx.To != to {
				continue
			}

			transaction, err := toTransactionResponse(ns, tx)
			if err != nil {
				http.Error(w, "failed to encode response", http.StatusInternalServerError)
				return
			}
			response = append(response, transaction)
		}

		if err := encode(w, r, response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func PostTransaction(s *state.State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx, err := decode[core.Transaction](r)
		if err != nil {
			http.Error(w, "failed to decode request payload", http.StatusBadRequest)
			return
		}

		err = s.UpsertTransaction(tx)
		if err != nil {
			http.Error(w, "failed to post transaction", http.StatusInternalServerError)
			return
		}

		response := struct {
			Status string `json:"status"`
		}{
			"transaction added to mempool",
		}

		if err := encode(w, r, response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func PostPeer(s *state.State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		peer, err := decode[peer.Peer](r)
		if err != nil {
			http.Error(w, "failed to decode request payload", http.StatusBadRequest)
			return
		}

		s.AddKnownPeer(peer)

		response := struct {
			Status string `json:"status"`
		}{
			"peer added to known peers",
		}

		if err := encode(w, r, response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func GetStatus(s *state.State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		latstBlock := s.LatestBlock()

		status := peer.PeerStatus{
			LatestBlockHash:   latstBlock.Hash(),
			LatestBlockNumber: latstBlock.Header.Number,
			KnownPeers:        s.KnownExternalPeers(),
		}

		if err := encode(w, r, status); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

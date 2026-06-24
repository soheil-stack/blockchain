package public

import (
	"net/http"

	"github.com/soheil-stack/blockchain/cmd/node/handlers"
	"github.com/soheil-stack/blockchain/internal/state"
)

func GetGenesis(s *state.State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		genesis := s.Genesis()
		_ = handlers.Encode(w, r, genesis)
	})
}

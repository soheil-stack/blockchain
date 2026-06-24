// Package state
package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
)

type StateConfig struct {
	Beneficiary    common.Address
	Genesis        core.Genesis
	EvHandler      EventHandler
	SelectStrategy string
}

type State struct {
	beneficiary common.Address
	genesis     core.Genesis
	db          *Database
	mempool     *Mempool
	evHandler   EventHandler
}

func NewState(config StateConfig) (*State, error) {
	evHandler := func(v string, args ...any) {
		if config.EvHandler != nil {
			config.EvHandler(v, args...)
		}
	}

	db := NewDatabase(config.Genesis, evHandler)
	mempool, err := NewMempool(config.SelectStrategy)
	if err != nil {
		return nil, err
	}

	return &State{
		beneficiary: config.Beneficiary,
		genesis:     config.Genesis,
		db:          db,
		mempool:     mempool,
		evHandler:   evHandler,
	}, nil
}

func (state *State) Shutdown() error {
	state.evHandler("state: shutdown started")
	defer state.evHandler("state: shutdown finished")

	return nil
}

func (state *State) Genesis() core.Genesis {
	return state.genesis
}

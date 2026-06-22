// Package state
package state

import "github.com/soheil-stack/blockchain/internal/core"

type StateConfig struct {
	Genesis        core.Genesis
	EvHandler      EventHandler
	SelectStrategy string
}

type State struct {
	db      *Database
	mempool *Mempool
}

func NewState(config StateConfig) (*State, error) {
	db := NewDatabase(config.Genesis, config.EvHandler)
	mempool, err := NewMempool(config.SelectStrategy)
	if err != nil {
		return nil, err
	}

	return &State{
		db:      db,
		mempool: mempool,
	}, nil
}

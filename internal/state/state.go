// Package state
package state

import "github.com/soheil-stack/blockchain/internal/core"

type StateConfig struct {
	Genesis   core.Genesis
	EvHandler EventHandler
}

type State struct {
	db      *Database
	mempool *Mempool
}

func NewState(config StateConfig) *State {
	db := NewDatabase(config.Genesis, config.EvHandler)
	mempool := NewMempool()

	return &State{
		db:      db,
		mempool: mempool,
	}
}

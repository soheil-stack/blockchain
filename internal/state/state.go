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

func (state *State) Accounts() map[common.Address]core.Account {
	return state.db.Copy()
}

func (state *State) Account(address common.Address) (core.Account, bool) {
	return state.db.Query(address)
}

func (state *State) Mempool() []core.Transaction {
	return state.mempool.PickBest()
}

func (state *State) UpsertTransaction(tx core.Transaction) error {
	if err := tx.Verify(state.genesis.ChainID); err != nil {
		return err
	}

	// TODO: update GasPrice and GasUnits

	return state.mempool.Upsert(tx)
}

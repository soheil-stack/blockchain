// Package state
package state

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
)

type StateConfig struct {
	Beneficiary    common.Address
	Genesis        core.Genesis
	EvHandler      core.EventHandler
	SelectStrategy string
}

type State struct {
	beneficiary common.Address
	genesis     core.Genesis
	db          *Database
	mempool     *Mempool
	evHandler   core.EventHandler
	Worker      *Worker
}

var ErrMempoolIsEmpty = errors.New("no transaction in the mempool")

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

	state.Worker.Shutdown()

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

func (state *State) MempoolPickBest(n ...int) []core.Transaction {
	return state.mempool.PickBest(n...)
}

func (state *State) MempoolUpsert(tx core.Transaction) error {
	return state.mempool.Upsert(tx)
}

func (state *State) MempoolLength() int {
	return state.mempool.Length()
}

func (state *State) UpsertTransaction(tx core.Transaction) error {
	if err := tx.Verify(state.genesis.ChainID); err != nil {
		return err
	}

	// TODO: update GasPrice and GasUnits

	if err := state.MempoolUpsert(tx); err != nil {
		return err
	}

	state.Worker.SignalStartMining()

	return nil
}

func (state *State) MineNewBlock(ctx context.Context) (core.Block, error) {
	defer state.evHandler("state: MineNewBlock: MINING: completed")

	state.evHandler("state: MineNewBlock: MINING: check mempool count")

	if state.MempoolLength() == 0 {
		return core.Block{}, ErrMempoolIsEmpty
	}

	transactions := state.MempoolPickBest(int(state.genesis.TransactionPerBlock))

	block, err := core.NewBlock(ctx, core.BlockConfig{
		Beneficiary:  state.beneficiary,
		Difficulty:   state.genesis.Difficulty,
		MiningReward: state.genesis.MiningReward,
		PrevBlock:    state.db.LatestBlock(),
		StateRoot:    state.db.HashState(),
		Transactions: transactions,
		EvHandler:    state.evHandler,
	})
	if err != nil {
		return core.Block{}, err
	}

	return block, nil
}

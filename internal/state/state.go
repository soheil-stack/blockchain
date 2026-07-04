// Package state
package state

import (
	"context"
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
)

type StateConfig struct {
	Beneficiary    common.Address
	Genesis        core.Genesis
	EvHandler      core.EventHandler
	SelectStrategy string
	Storage        Storage
}

type State struct {
	beneficiary common.Address
	genesis     core.Genesis
	db          *Database
	mempool     *Mempool
	evHandler   core.EventHandler
	mu          sync.RWMutex
	Worker      *Worker
}

var ErrMempoolIsEmpty = errors.New("no transaction in the mempool")

func NewState(config StateConfig) (*State, error) {
	evHandler := func(v string, args ...any) {
		if config.EvHandler != nil {
			config.EvHandler(v, args...)
		}
	}

	db, err := NewDatabase(config.Genesis, config.Storage, evHandler)
	if err != nil {
		return nil, err
	}

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

	if err := state.ValidateBlockAndUpdateDatabase(block); err != nil {
		return core.Block{}, err
	}

	return block, nil
}

func (state *State) ValidateBlockAndUpdateDatabase(block core.Block) error {
	state.mu.Lock()
	defer state.mu.Unlock()

	state.evHandler("state: ValidateBlockAndUpdateDatabase: validate block")

	if err := block.Validate(state.db.LatestBlock(), state.db.HashState(), state.evHandler); err != nil {
		return err
	}

	state.evHandler("state: ValidateBlockAndUpdateDatabase: write block")

	if err := state.db.Write(block); err != nil {
		return err
	}
	state.db.SetLatestBlock(block)

	state.evHandler("state ValidateBlockAndUpdateDatabase: update accounts")

	for _, tx := range block.Transactions {
		state.evHandler("state: ValidateBlockAndUpdateDatabase: update accounts: tx[%s]", tx)

		state.mempool.Remove(tx)

		if err := state.db.ApplyTransaction(block, tx); err != nil {
			state.evHandler("state: ValidateBlockAndUpdateDatabase: update accounts: ERROR: %w", err)
			continue
		}
	}

	state.evHandler("state: ValidateBlockAndUpdateDatabase: apply mining reward")

	state.db.ApplyMiningReward(block)

	return nil
}

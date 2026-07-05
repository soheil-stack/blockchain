// Package state
package state

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/soheil-stack/blockchain/internal/peer"
)

type StateConfig struct {
	Beneficiary    common.Address
	Genesis        core.Genesis
	EvHandler      core.EventHandler
	SelectStrategy string
	Storage        Storage
	KnownPeers     *peer.PeerSet
}

type State struct {
	beneficiary common.Address
	genesis     core.Genesis
	db          *Database
	mempool     *Mempool
	evHandler   core.EventHandler
	mu          sync.RWMutex
	knownPeers  *peer.PeerSet
	host        string
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
		knownPeers:  config.KnownPeers,
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

	// TODO: do not hardcode gasUnits
	tx.GasUnits = 10
	tx.GasPrice = state.genesis.GasPrice

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
			state.evHandler("state: ValidateBlockAndUpdateDatabase: update accounts: ERROR: %s", err)
			continue
		}
	}

	state.evHandler("state: ValidateBlockAndUpdateDatabase: apply mining reward")

	state.db.ApplyMiningReward(block)

	return nil
}

func (state *State) AddKnownPeer(peer peer.Peer) bool {
	return state.knownPeers.Add(peer)
}

func (state *State) RemoveKnownPeer(peer peer.Peer) {
	state.knownPeers.Remove(peer)
}

func (state *State) KnownExternalPeers() []peer.Peer {
	return state.knownPeers.Copy(state.host)
}

func (state *State) KnownPeers() []peer.Peer {
	return state.knownPeers.Copy("")
}

func (state *State) LatestBlock() core.Block {
	return state.db.LatestBlock()
}

func (state *State) Host() string {
	return state.host
}

func (state *State) NetRequestPeerStatus(p peer.Peer) (peer.PeerStatus, error) {
	state.evHandler("state: NetRequestPeerStatus: started: %s", p.Host)
	defer state.evHandler("state: NetRequestPeerStatus: completed: %s", p.Host)

	var peerStatus peer.PeerStatus
	url := fmt.Sprintf("http://%s/node/status", p.Host)
	err := core.Send(http.MethodGet, url, nil, &peerStatus)
	if err != nil {
		return peer.PeerStatus{}, err
	}

	state.evHandler("state: NetRequestPeerStatus: peer-node[%s]: latest-blockNumber[%d]: peer-list[%s]", p.Host, peerStatus.LatestBlockNumber, peerStatus.KnownPeers)

	return peerStatus, nil
}

func (state *State) NetRequestPeerMempool(p peer.Peer) ([]core.Transaction, error) {
	state.evHandler("state: NetRequestPeerMempool: started: %s", p.Host)
	defer state.evHandler("state: NetRequestPeerMempool: completed: %s", p.Host)

	var mempool []core.Transaction
	url := fmt.Sprintf("http://%s/mempool/transactions", p.Host)
	err := core.Send(http.MethodGet, url, nil, &mempool)
	if err != nil {
		return nil, err
	}

	state.evHandler("state: NetRequestPeerMempool: len[%d]", len(mempool))

	return mempool, err
}

func (state *State) NetRequestPeerBlocks(p peer.Peer) error {
	return nil
}

func (state *State) NetSendNodeAvailableToPeers() error {
	state.evHandler("state: NetSendNodeAvailableToPeers: started")
	defer state.evHandler("state: NetSendNodeAvailableToPeers: completed")

	host := state.Host()
	selfPeer := peer.Peer{
		Host: host,
	}

	for _, peer := range state.KnownExternalPeers() {
		state.evHandler("state: NetSendNodeAvailableToPeers: send: host[%s] to peer[%s]", host, peer)

		url := fmt.Sprintf("http://%s/node/peers", peer.Host)
		if err := core.Send(http.MethodPost, url, selfPeer, nil); err != nil {
			state.evHandler("state: NetSendNodeAvailableToPeers: ERROR: %s", err)
		}
	}

	return nil
}

func (state *State) NetSendBlockToPeers(block core.Block) error {
	return nil
}

func (state *State) NetSendTxToPeers(tx core.Transaction) error {
	return nil
}

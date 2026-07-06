package state

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/soheil-stack/blockchain/internal/core"
)

type Worker struct {
	state        *State
	wg           sync.WaitGroup
	shut         chan struct{}
	startMining  chan bool
	cancelMining chan bool
	txSharing    chan core.Transaction
	evHandler    core.EventHandler
}

func RunWorker(state *State, evHandler core.EventHandler) {
	worker := Worker{
		state:        state,
		shut:         make(chan struct{}),
		startMining:  make(chan bool, 1),
		cancelMining: make(chan bool, 1),
		txSharing:    make(chan core.Transaction, 100),
		evHandler:    evHandler,
	}

	state.Worker = &worker

	worker.Sync()

	operations := []func(){
		worker.shareTxOperation,
		worker.powOperation,
	}

	worker.wg.Add(len(operations))

	for _, op := range operations {
		go func(op func()) {
			defer worker.wg.Done()
			op()
		}(op)
	}
}

func (worker *Worker) Shutdown() {
	worker.evHandler("worker: shutdown: started")
	defer worker.evHandler("worker: shutdown: completed")

	worker.evHandler("worker: shutdown: signal cancel mining")
	worker.SignalCancelMining()

	worker.evHandler("worker: shutdown: terminate goroutines")
	close(worker.shut)
	worker.wg.Wait()
}

func (worker *Worker) Sync() {
	worker.evHandler("worker: Sync: started")
	defer worker.evHandler("worker: Sync: completed")

	for _, peer := range worker.state.KnownExternalPeers() {
		peerStatus, err := worker.state.NetRequestPeerStatus(peer)
		if err != nil {
			worker.evHandler("worker: Sync: queryPeerStatus: %s: ERROR: %s", peer, err)
			continue
		}

		worker.addNewPeers(peerStatus.KnownPeers)

		mempool, err := worker.state.NetRequestPeerMempool(peer)
		if err != nil {
			worker.evHandler("worker: Sync: retrievePeerMempool: %s: ERROR: %s", peer, err)
			continue
		}

		for _, tx := range mempool {
			worker.evHandler("worker: Sync: retrievePeerMempool: %s: Add tx[%s]", peer, tx)
			_ = worker.state.MempoolUpsert(tx)
		}

		if peerStatus.LatestBlockNumber > worker.state.LatestBlock().Header.Number {
			worker.evHandler("worker: Sync: retrievePeerBlocks: %s: latestBlockNumber[%d]", peer, peerStatus.LatestBlockNumber)

			if err := worker.state.NetRequestPeerBlocks(peer); err != nil {
				worker.evHandler("worker: Sync: retrievePeerBlocks: %s: ERROR: %s", peer, err)
			}
		}
	}

	worker.state.NetSendNodeAvailableToPeers()
}

func (worker *Worker) SignalStartMining() {
	select {
	case worker.startMining <- true:
	default:
	}
	worker.evHandler("worker: SignalStartMining: MINING: signaled")
}

func (worker *Worker) SignalCancelMining() {
	select {
	case worker.cancelMining <- true:
	default:
	}
	worker.evHandler("worker: SignalCancelMining: MINING: CANCEL: signaled")
}

func (worker *Worker) SignalShareTx(tx core.Transaction) {
	select {
	case worker.txSharing <- tx:
		worker.evHandler("worker: SignalShareTx: share transaction signaled")
	default:
		worker.evHandler("worker: SignalShareTx: queue is full, transaction won't be shared")
	}
}

func (worker *Worker) powOperation() {
	worker.evHandler("worker: powOperation: G started")
	defer worker.evHandler("worker: powOperation: G completed")

	for {
		select {
		case <-worker.startMining:
			if !worker.isShutdown() {
				worker.runPowOperation()
			}
		case <-worker.shut:
			worker.evHandler("worker: powOperation: received shutdown signal")
			return
		}
	}
}

func (worker *Worker) runPowOperation() {
	worker.evHandler("worker: runPowOperation: MINING: started")
	defer worker.evHandler("worker: runPowOperation: MINING: completed")

	if worker.state.MempoolLength() == 0 {
		worker.evHandler("worker: runPowOperation: MINING: no transaction in the mempool")
		return
	}

	defer func() {
		length := worker.state.MempoolLength()
		if length > 0 {
			worker.evHandler("worker: runPowOperation: MINING: signal new mining operation: Txs[%d]", length)
			worker.SignalStartMining()
		}
	}()

	select {
	case <-worker.cancelMining:
		worker.evHandler("worker: runPowOperation: MINING: drained cancel channel")
	default:
	}

	var wg sync.WaitGroup
	wg.Add(2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer func() {
			wg.Done()
			cancel()
		}()

		select {
		case <-worker.cancelMining:
			worker.evHandler("worker: runPowOperation: MINING: cancel mining")
		case <-ctx.Done():
		}
	}()

	go func() {
		defer func() {
			wg.Done()
			cancel()
		}()

		t := time.Now()
		block, err := worker.state.MineNewBlock(ctx)
		duration := time.Since(t)

		worker.evHandler("worker: runPowOperation: MINING: completed: duration[%f]", duration.Seconds())

		if err != nil {
			switch {
			case errors.Is(err, ErrMempoolIsEmpty):
				worker.evHandler("worker: runPowOperation: MINING: no transaction in the mempool")
			case ctx.Err() != nil:
				worker.evHandler("worker: runPowOperation: MINING: CANCEL: completed")
			default:
				worker.evHandler("worker: runPowOperation: MINING: ERROR: %s", err)
			}

			return
		}

		// we mined a block
		worker.state.NetSendBlockToPeers(block)
	}()

	wg.Wait()
}

func (worker *Worker) shareTxOperation() {
	worker.evHandler("worker: shareTxOperation: G started")
	defer worker.evHandler("worker: shareTxOperation: G completed")

	for {
		select {
		case tx := <-worker.txSharing:
			if !worker.isShutdown() {
				worker.state.NetSendTxToPeers(tx)
			}
		case <-worker.shut:
			worker.evHandler("worker: shareTxOperation: received shut signal")
			return
		}
	}
}

func (worker *Worker) addNewPeers(peers []core.Peer) {
	worker.evHandler("worker: runPeerUpdatesOperation: addNewPeers: started")
	defer worker.evHandler("worker: runPeerUpdatesOperation: addNewPeers: completed")

	for _, peer := range peers {
		if peer.Match(worker.state.Host()) {
			continue
		}

		if worker.state.AddKnownPeer(peer) {
			worker.evHandler("worker: runPeerUpdatesOperation: addNewPeers: add peer nodes: adding peer-node %s", peer)
		}
	}
}

func (worker *Worker) isShutdown() bool {
	select {
	case <-worker.shut:
		return true
	default:
		return false
	}
}

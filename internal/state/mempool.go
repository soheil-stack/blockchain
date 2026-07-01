package state

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
)

type Mempool struct {
	mu           sync.RWMutex
	selectorFn   SelectorFn
	Transactions map[string]core.Transaction
}

func NewMempool(strategy string) (*Mempool, error) {
	selectorFn, err := Selector(strategy)
	if err != nil {
		return nil, err
	}

	return &Mempool{
		selectorFn:   selectorFn,
		Transactions: make(map[string]core.Transaction),
	}, nil
}

func (m *Mempool) Length() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.Transactions)
}

func (m *Mempool) Upsert(tx core.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := txKey(tx)

	if oldTx, ok := m.Transactions[key]; ok {
		if tx.Tip < uint64(float64(oldTx.Tip)*1.10) {
			return errors.New("replacing a transaction requires a 10% bump in the tip")
		}
	}

	m.Transactions[key] = tx

	return nil
}

func (m *Mempool) Remove(tx core.Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := txKey(tx)
	delete(m.Transactions, key)
}

func (m *Mempool) Truncate() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Transactions = make(map[string]core.Transaction)
}

func (m *Mempool) PickBest(n ...int) []core.Transaction {
	number := len(m.Transactions)
	if len(n) > 0 {
		number = n[0]
	}
	m.mu.RLock()
	accountTxs := make(map[common.Address][]core.Transaction)
	for _, tx := range m.Transactions {
		accountTxs[tx.From] = append(accountTxs[tx.From], tx)
	}
	m.mu.RUnlock()

	return m.selectorFn(accountTxs, number)
}

func txKey(tx core.Transaction) string {
	return fmt.Sprintf("%d:%s", tx.Nonce, tx.From)
}

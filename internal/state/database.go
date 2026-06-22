package state

import (
	"maps"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
)

type EventHandler func(v string, args ...any)

type Database struct {
	mu       sync.RWMutex
	accounts map[common.Address]core.Account
}

func NewDatabase(genesis core.Genesis, evHandler EventHandler) *Database {
	db := Database{
		accounts: make(map[common.Address]core.Account),
	}

	for addressHex, balance := range genesis.Balances {
		address := common.HexToAddress(addressHex)
		db.accounts[address] = core.NewAccount(address, balance)
	}

	return &db
}

func (db *Database) Remove(address common.Address) {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.accounts, address)
}

func (db *Database) Query(address common.Address) (core.Account, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	account, ok := db.accounts[address]
	return account, ok
}

func (db *Database) Copy() map[common.Address]core.Account {
	db.mu.RLock()
	defer db.mu.RUnlock()

	accounts := make(map[common.Address]core.Account)
	maps.Copy(accounts, db.accounts)

	return accounts
}

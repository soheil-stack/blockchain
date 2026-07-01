package state

import (
	"crypto/sha256"
	"encoding/json"
	"maps"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
)

type Database struct {
	mu          sync.RWMutex
	accounts    map[common.Address]core.Account
	latestBlock core.Block
}

func NewDatabase(genesis core.Genesis, evHandler core.EventHandler) *Database {
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

func (db *Database) SetLatestBlock(block core.Block) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.latestBlock = block
}

func (db *Database) LatestBlock() core.Block {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.latestBlock
}

func (db *Database) HashState() common.Hash {
	accounts := make([]core.Account, 0, len(db.accounts))
	db.mu.RLock()
	for _, account := range db.accounts {
		accounts = append(accounts, account)
	}
	db.mu.RUnlock()

	sort.Sort(core.ByAccount(accounts))

	data, err := json.Marshal(accounts)
	if err != nil {
		return common.Hash{}
	}

	return sha256.Sum256(data)
}

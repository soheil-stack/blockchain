package state

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"maps"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
)

type Storage interface {
	Write(block core.Block) error
	GetBlock(number int) (core.Block, error)
	ForEach() Iterator
	Close() error
	Reset() error
}

type Iterator interface {
	Next() (core.Block, error)
	Done() bool
}

type Database struct {
	mu          sync.RWMutex
	accounts    map[common.Address]core.Account
	latestBlock core.Block
	storage     Storage
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

func (db *Database) ApplyMiningReward(block core.Block) {
	db.mu.Lock()
	defer db.mu.Unlock()

	account, exists := db.accounts[block.Header.Beneficiary]
	if !exists {
		account = core.NewAccount(block.Header.Beneficiary, 0)
	}

	account.Balance += block.Header.MiningReward
	db.accounts[block.Header.Beneficiary] = account
}

func (db *Database) ApplyTransaction(block core.Block, tx core.Transaction) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	from, exists := db.accounts[tx.From]
	if !exists {
		from = core.NewAccount(tx.From, 0)
	}

	to, exists := db.accounts[tx.To]
	if !exists {
		to = core.NewAccount(tx.To, 0)
	}

	beneficiary, exists := db.accounts[block.Header.Beneficiary]
	if !exists {
		beneficiary = core.NewAccount(block.Header.Beneficiary, 0)
	}

	gasFee := min(tx.GasPrice*tx.GasUnits, from.Balance)
	from.Balance -= gasFee
	beneficiary.Balance += gasFee

	db.accounts[tx.From] = from
	db.accounts[block.Header.Beneficiary] = beneficiary

	if tx.Nonce != (from.Nonce + 1) {
		return fmt.Errorf("transaction invalid, wrong nonce, got %d, expected %d", tx.Nonce, from.Nonce+1)
	}

	if from.Balance < (tx.Value + tx.Tip) {
		return fmt.Errorf("transaction invalid, insufficient funds, balance %d, needed %d", from.Balance, (tx.Value + tx.Tip))
	}

	from.Balance -= tx.Value
	to.Balance += tx.Value

	from.Balance -= tx.Tip
	beneficiary.Balance += tx.Tip

	from.Nonce = tx.Nonce

	db.accounts[tx.From] = from
	db.accounts[tx.To] = to
	db.accounts[block.Header.Beneficiary] = beneficiary

	return nil
}

func (db *Database) Write(block core.Block) error {
	return db.storage.Write(block)
}

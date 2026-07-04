package core

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

var ErrChainForked = errors.New("blockchain forked")

type EventHandler func(v string, args ...any)

type BlockHeader struct {
	Number           uint64         `json:"number"`
	PrevBlockHash    common.Hash    `json:"prevBlockHash"`
	Timestamp        uint64         `json:"timestamp"`
	Beneficiary      common.Address `json:"beneficiary"`
	Difficulty       uint16         `json:"difficulty"`
	MiningReward     uint64         `json:"miningReward"`
	StateRoot        common.Hash    `json:"stateRoot"`
	TransactionsRoot common.Hash    `json:"transactionsRoot"`
	Nonce            uint64         `json:"nonce"`
}

type BlockConfig struct {
	Beneficiary  common.Address
	Difficulty   uint16
	MiningReward uint64
	PrevBlock    Block
	StateRoot    common.Hash
	Transactions []Transaction
	EvHandler    EventHandler
}

type Block struct {
	Header       BlockHeader   `json:"header"`
	Transactions []Transaction `json:"transactions"`
}

func NewBlock(ctx context.Context, config BlockConfig) (Block, error) {
	var prevBlockHash common.Hash
	if config.PrevBlock.Header.Number > 0 {
		prevBlockHash = config.PrevBlock.Hash()
	}

	tree := NewMarkleTree(config.Transactions)

	block := Block{
		Header: BlockHeader{
			Number:           config.PrevBlock.Header.Number + 1,
			PrevBlockHash:    prevBlockHash,
			Timestamp:        uint64(time.Now().UTC().UnixMilli()),
			Beneficiary:      config.Beneficiary,
			Difficulty:       config.Difficulty,
			MiningReward:     config.MiningReward,
			StateRoot:        config.StateRoot,
			TransactionsRoot: tree.MerkleRoot(),
			Nonce:            0,
		},
		Transactions: config.Transactions,
	}

	if err := block.performPOW(ctx, config.EvHandler); err != nil {
		return Block{}, err
	}

	return block, nil
}

func (block Block) Validate(previousBlock Block, stateRoot common.Hash, evHandler EventHandler) error {
	nextNumber := previousBlock.Header.Number + 1

	evHandler("block: Validate: block[%d]: check: block number is the next number", block.Header.Number)

	if block.Header.Number != nextNumber {
		return fmt.Errorf("block number[%d] is not the next number[%d]", block.Header.Number, nextNumber)
	}

	evHandler("block: Validate: block[%d]: check: chain is not forked", block.Header.Number)

	if block.Header.Number >= (nextNumber + 2) {
		return ErrChainForked
	}

	evHandler("block: Validate: block[%d]: check: block difficulty is the same or greater than parent block difficulty", block.Header.Number)

	if block.Header.Difficulty < previousBlock.Header.Difficulty {
		return fmt.Errorf("block difficulty[%d] is less than previous block difficulty[%d]", block.Header.Difficulty, previousBlock.Header.Difficulty)
	}

	evHandler("block: Validate: block[%d]: check: block hash has been solved", block.Header.Number)

	hash := block.Hash()
	if !isHashSolved(hash, block.Header.Difficulty) {
		return fmt.Errorf("%s invalid block hash", hash)
	}

	evHandler("block: Validate: block[%d]: check: parent hash does match parent block", block.Header.Number)

	if block.Header.PrevBlockHash != previousBlock.Hash() {
		return fmt.Errorf("parent block hash[%s] does not match our known parent block hash[%s]", block.Header.PrevBlockHash, previousBlock.Hash())
	}

	evHandler("block: Validate: block[%d]: check: block's timestamp is greater than parent block's timestamp", block.Header.Number)

	if block.Header.Timestamp <= previousBlock.Header.Timestamp {
		return fmt.Errorf("block timestamp[%d] is before parent block timestamp[%d]", block.Header.Timestamp, previousBlock.Header.Timestamp)
	}

	evHandler("block: Validate: block[%d]: check: state root hash does match current database", block.Header.Number)

	if block.Header.StateRoot != stateRoot {
		return fmt.Errorf("state root hash[%s] does not match current database hash[%s]", block.Header.StateRoot, stateRoot)
	}

	evHandler("block: Validate: block[%d]: check: merkle root does match transactions", block.Header.Number)

	tree := NewMarkleTree(block.Transactions)
	if block.Header.TransactionsRoot != tree.MerkleRoot() {
		return fmt.Errorf("merkle root[%s] does not match transactions[%s]", block.Header.TransactionsRoot, tree.MerkleRoot())
	}

	return nil
}

func (block Block) performPOW(ctx context.Context, ev EventHandler) error {
	ev("block: Perfrom POW: MINING: started")
	defer ev("block: Perfrom POW: MINING: completed")

	for _, tx := range block.Transactions {
		ev("block: Perform POW: MINING: tx[%s]", tx)
	}

	block.Header.Nonce = rand.Uint64()

	ev("block PerformPOW: MINING: running")

	var attempts uint64
	for {
		attempts++
		if attempts%1_000_000 == 0 {
			ev("block: Perfrom POW: MINING: running: attempts[%d]", attempts)
		}

		if ctx.Err() != nil {
			ev("block: Perfrom POW: MINING: CANCELLED")
			return ctx.Err()
		}

		hash := block.Hash()
		if !isHashSolved(hash, block.Header.Difficulty) {
			block.Header.Nonce++
			continue
		}

		ev("block: Perform POW: MINING: SOLVED: prevBlockHash[%s]: newBlockHash[%s]", block.Header.PrevBlockHash, hash)

		return nil
	}
}

func (block Block) Hash() common.Hash {
	data, err := json.Marshal(block.Header)
	if err != nil {
		return common.Hash{}
	}

	return sha256.Sum256(data)
}

func isHashSolved(hash common.Hash, difficulty uint16) bool {
	const match = "0x00000000000000000"

	difficulty += 2
	return hash.Hex()[:difficulty] == match[:difficulty]
}

package core

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

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

func (b *Block) performPOW(ctx context.Context, ev EventHandler) error {
	ev("block: Perfrom POW: MINING: started")
	defer ev("block: Perfrom POW: MINING: completed")

	for _, tx := range b.Transactions {
		ev("block: Perform POW: MINING: tx[%s]", tx)
	}

	b.Header.Nonce = rand.Uint64()

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

		hash := b.Hash()
		if !isHashSolved(hash, b.Header.Difficulty) {
			b.Header.Nonce++
			continue
		}

		ev("block: Perform POW: MINING: SOLVED: prevBlockHash[%s]: newBlockHash[%s]", b.Header.PrevBlockHash, hash)

		return nil
	}
}

func (b Block) Hash() common.Hash {
	data, err := json.Marshal(b.Header)
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

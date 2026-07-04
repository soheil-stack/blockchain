package core

import (
	_ "embed"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

//go:embed data/genesis.json
var genesisBytes []byte

type Genesis struct {
	Date                time.Time                 `json:"date"`
	ChainID             uint64                    `json:"chainID"`
	TransactionPerBlock uint64                    `json:"transactionPerBlock"`
	Difficulty          uint16                    `json:"difficulty"`
	MiningReward        uint64                    `json:"miningReward"`
	GasPrice            uint64                    `json:"gasPrice"`
	Balances            map[common.Address]uint64 `json:"balances"`
}

func LoadGenesis() (Genesis, error) {
	var genesis Genesis

	err := json.Unmarshal(genesisBytes, &genesis)
	if err != nil {
		return Genesis{}, err
	}

	return genesis, nil
}

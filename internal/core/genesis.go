package core

import (
	"encoding/json"
	"os"
	"time"
)

type Genesis struct {
	Date                time.Time         `json:"date"`
	ChainID             uint64            `json:"chainID"`
	TransactionPerBlock uint64            `json:"transactionPerBlock"`
	Difficulty          uint64            `json:"difficulty"`
	MiningReward        uint64            `json:"miningReward"`
	GasPrice            uint64            `json:"gasPrice"`
	Balances            map[string]uint64 `json:"balances"`
}

func LoadGenesis() (Genesis, error) {
	var genesis Genesis

	data, err := os.ReadFile("zblock/genesis.json")
	if err != nil {
		return Genesis{}, err
	}

	err = json.Unmarshal(data, &genesis)
	if err != nil {
		return Genesis{}, err
	}

	return genesis, nil
}

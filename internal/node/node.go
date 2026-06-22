// Package node
package node

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/soheil-stack/blockchain/internal/state"
)

type Config struct {
	Beneficiary common.Address
	Genesis     core.Genesis
	EvHandler   state.EventHandler
}

type Node struct {
	State     *state.State
	EvHandler state.EventHandler
}

func New(config Config) *Node {
	ev := func(v string, args ...any) {
		if config.EvHandler != nil {
			config.EvHandler(v, args...)
		}
	}

	state := state.NewState(state.StateConfig{
		Genesis:   config.Genesis,
		EvHandler: ev,
	})

	return &Node{
		State:     state,
		EvHandler: ev,
	}
}

func (node *Node) Shutdown() error {
	node.EvHandler("node: shutdown started")
	defer node.EvHandler("node: shutdown finished")

	return nil
}

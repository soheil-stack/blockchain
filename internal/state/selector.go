package state

import (
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
)

type SelectorFn func(map[common.Address][]core.Transaction, int) []core.Transaction

const (
	StrategyTip = "tip"
)

func tipSelector(accountTxs map[common.Address][]core.Transaction, n int) []core.Transaction {
	for address := range accountTxs {
		sort.Sort(byNonce(accountTxs[address]))
	}

	var final []core.Transaction
	for len(final) < n {
		var row []core.Transaction
		for address := range accountTxs {
			if len(accountTxs[address]) > 0 {
				row = append(row, accountTxs[address][0])
				accountTxs[address] = accountTxs[address][1:]
			}
		}

		if row == nil {
			break
		}

		need := n - len(final)
		if len(row) > need {
			sort.Sort(byTip(row))
			final = append(final, row[:need]...)
			break
		}
		final = append(final, row...)
	}

	return final
}

var strategies = map[string]SelectorFn{
	StrategyTip: tipSelector,
}

func Selector(strategy string) (SelectorFn, error) {
	selectorFn, ok := strategies[strategy]
	if !ok {
		return nil, fmt.Errorf("unknown selector strategy: %s", strategy)
	}

	return selectorFn, nil
}

type byNonce []core.Transaction

func (bn byNonce) Len() int {
	return len(bn)
}

func (bn byNonce) Less(i, j int) bool {
	return bn[i].Nonce < bn[j].Nonce
}

func (bn byNonce) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}

type byTip []core.Transaction

func (bt byTip) Len() int {
	return len(bt)
}

func (bt byTip) Less(i, j int) bool {
	return bt[i].Tip > bt[j].Tip
}

func (bt byTip) Swap(i, j int) {
	bt[i], bt[j] = bt[j], bt[i]
}

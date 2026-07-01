package core

import "github.com/ethereum/go-ethereum/common"

type Account struct {
	Address common.Address
	Nonce   uint64
	Balance uint64
}

func NewAccount(address common.Address, balance uint64) Account {
	return Account{
		Address: address,
		Balance: balance,
	}
}

type ByAccount []Account

func (ba ByAccount) Len() int {
	return len(ba)
}

func (ba ByAccount) Less(i, j int) bool {
	return ba[i].Address.Hex() < ba[j].Address.Hex()
}

func (ba ByAccount) Swap(i, j int) {
	ba[i], ba[j] = ba[j], ba[i]
}

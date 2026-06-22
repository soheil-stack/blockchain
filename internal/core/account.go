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

package server

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/soheil-stack/blockchain/internal/nameservice"
)

type AccountResponse struct {
	Address common.Address `json:"address"`
	Name    string         `json:"name"`
	Balance uint64         `json:"balance"`
	Nonce   uint64         `json:"nonce"`
}

func toAccountResponse(ns *nameservice.NameService, account core.Account) AccountResponse {
	return AccountResponse{
		Address: account.Address,
		Name:    ns.Get(account.Address),
		Balance: account.Balance,
		Nonce:   account.Nonce,
	}
}

type TransactionResponse struct {
	ChainID     uint64         `json:"chainID"`
	Nonce       uint64         `json:"nonce"`
	FromAddress common.Address `json:"from"`
	FromName    string         `json:"fromName"`
	To          common.Address `json:"to"`
	ToName      string         `json:"toName"`
	Value       uint64         `json:"value"`
	Tip         uint64         `json:"tip"`
	Data        []byte         `json:"data"`
	Sig         []byte         `json:"sig"`
	GasPrice    uint64         `json:"gasPrice"`
	GasUnits    uint64         `json:"gasUnits"`
}

func toTransactionResponse(ns *nameservice.NameService, tx core.Transaction) (TransactionResponse, error) {
	sig, err := tx.Signature()
	if err != nil {
		return TransactionResponse{}, err
	}

	return TransactionResponse{
		ChainID:     tx.ChainID,
		Nonce:       tx.Nonce,
		FromAddress: tx.From,
		FromName:    ns.Get(tx.From),
		To:          tx.To,
		ToName:      ns.Get(tx.To),
		Value:       tx.Value,
		Tip:         tx.Tip,
		Data:        tx.Data,
		Sig:         sig,
		GasPrice:    tx.GasPrice,
		GasUnits:    tx.GasUnits,
	}, nil
}

package core

import (
	"io/fs"
	"maps"
	"path"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type NameService struct {
	accounts map[common.Address]string
}

func NewNameService(root string) (*NameService, error) {
	accounts := make(map[common.Address]string)

	fn := func(file string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path.Ext(file) != ".ecdsa" {
			return nil
		}

		privateKey, err := crypto.LoadECDSA(file)
		if err != nil {
			return err
		}

		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		accounts[address] = strings.TrimSuffix(path.Base(file), ".ecdsa")

		return nil
	}

	if err := filepath.Walk(root, fn); err != nil {
		return nil, err
	}

	return &NameService{
		accounts: accounts,
	}, nil
}

func (ns *NameService) Get(address common.Address) string {
	name, ok := ns.accounts[address]
	if !ok {
		return address.Hex()
	}

	return name
}

func (ns *NameService) Copy() map[common.Address]string {
	accounts := make(map[common.Address]string)
	maps.Copy(accounts, ns.accounts)

	return accounts
}

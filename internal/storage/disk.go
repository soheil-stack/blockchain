// Package storage
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/ethereum/go-ethereum/common"
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/soheil-stack/blockchain/internal/state"
)

type Disk struct {
	dbPath string
}

func NewDisk(dbPath string) (*Disk, error) {
	if err := os.MkdirAll(dbPath, 0o755); err != nil {
		return nil, err
	}

	return &Disk{
		dbPath: dbPath,
	}, nil
}

func (disk *Disk) Close() error {
	return nil
}

func (disk *Disk) Write(block core.Block) error {
	blockData := struct {
		Hash         common.Hash        `json:"hash"`
		Header       core.BlockHeader   `json:"header"`
		Transactions []core.Transaction `json:"transactions"`
	}{
		block.Hash(),
		block.Header,
		block.Transactions,
	}

	data, err := json.MarshalIndent(blockData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(disk.getPath(block.Header.Number), data, 0o755)
}

func (disk *Disk) GetBlock(number uint64) (core.Block, error) {
	data, err := os.ReadFile(disk.getPath(number))
	if err != nil {
		return core.Block{}, err
	}

	var block core.Block
	err = json.Unmarshal(data, &block)

	return block, err
}

func (disk *Disk) Reset() error {
	if err := os.RemoveAll(disk.dbPath); err != nil {
		return err
	}

	return os.MkdirAll(disk.dbPath, 0o755)
}

func (disk *Disk) ForEach() state.Iterator {
	return &diskIterator{storage: disk}
}

func (disk *Disk) getPath(blockNumber uint64) string {
	return path.Join(disk.dbPath, fmt.Sprintf("%d.json", blockNumber))
}

type diskIterator struct {
	storage *Disk
	current uint64
	efc     bool
}

func (it *diskIterator) Next() (core.Block, error) {
	if it.Done() {
		return core.Block{}, errors.New("end of chain")
	}

	it.current++
	block, err := it.storage.GetBlock(it.current)
	if errors.Is(err, os.ErrNotExist) {
		it.efc = true
	}

	return block, err
}

func (it *diskIterator) Done() bool {
	return it.efc
}

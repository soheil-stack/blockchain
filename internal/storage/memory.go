package storage

import (
	"errors"
	"sync"

	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/soheil-stack/blockchain/internal/state"
)

var ErrBlockNotFound = errors.New("block not found")

type Memory struct {
	mu     sync.RWMutex
	blocks []core.Block
}

func NewMemory() *Memory {
	return &Memory{}
}

func (memory *Memory) Close() error {
	return nil
}

func (memory *Memory) Write(block core.Block) error {
	memory.mu.Lock()
	defer memory.mu.Unlock()

	length := len(memory.blocks)
	if length+1 != int(block.Header.Number) {
		return errors.New("block is out of order")
	}

	memory.blocks = append(memory.blocks, block)
	return nil
}

func (memory *Memory) GetBlock(number uint64) (core.Block, error) {
	memory.mu.RLock()
	defer memory.mu.RUnlock()

	if number == 0 || number > uint64(len(memory.blocks)) {
		return core.Block{}, ErrBlockNotFound
	}

	return memory.blocks[number-1], nil
}

func (memory *Memory) Reset() error {
	memory.mu.Lock()
	defer memory.mu.Unlock()

	memory.blocks = make([]core.Block, 0)
	return nil
}

func (memory *Memory) ForEach() state.Iterator {
	return &memoryIterator{storage: memory}
}

type memoryIterator struct {
	storage *Memory
	current uint64
	eoc     bool
}

func (it *memoryIterator) Next() (core.Block, error) {
	if it.Done() {
		return core.Block{}, errors.New("end of chain")
	}

	it.current++
	block, err := it.storage.GetBlock(it.current)
	if errors.Is(err, ErrBlockNotFound) {
		it.eoc = true
	}

	return block, err
}

func (it *memoryIterator) Done() bool {
	return it.eoc
}

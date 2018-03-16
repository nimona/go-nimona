package crdt

import (
	"errors"
	"sync"
)

var (
	ErrorMissingHash  = errors.New("Block is missing its hash")
	ErrorMissingBlock = errors.New("Block is missing")
)

// BlockStore is a temporary place to store blocks hashgraph blocks
// TODO Partition per thread
// TODO Make thread safe
// TODO Allow for persistence
// TODO Make into interface
type BlockStore struct {
	sync.RWMutex
	blocks map[string]*Block
}

func (store *BlockStore) Add(block *Block) error {
	if block.Hash == "" {
		return ErrorMissingHash
	}
	store.Lock()
	defer store.Unlock()
	if _, ok := store.blocks[block.Hash]; ok {
		return nil
	}
	store.blocks[block.Hash] = block
	return nil
}

func (store *BlockStore) Get(blockID string) (*Block, error) {
	store.RLock()
	defer store.RUnlock()
	return store.get(blockID)
}

func (store *BlockStore) get(blockID string) (*Block, error) {
	block, ok := store.blocks[blockID]
	if !ok {
		return nil, ErrorMissingBlock
	}
	return block, nil
}

func (store *BlockStore) FindSubscribers(blockID string) ([]string, error) {
	store.RLock()
	defer store.RUnlock()

	_, err := store.get(blockID)
	if err != nil {
		return nil, err
	}

	subscribers := []string{}
	children := store.findChildrenFlat(blockID)
	for _, child := range children {
		cb := store.blocks[child]
		if cb.Event.Type == EventTypeGraphSubscribe {
			subscribers = append(subscribers, cb.Event.Data)
		}
		subscribers = append(subscribers, cb.Event.Author)
	}
	return SliceUniq(subscribers), nil
}

func (store *BlockStore) findChildrenFlat(blockID string) []string {
	children := store.findChildren(blockID)
	for _, child := range children {
		cchildren := store.findChildren(child)
		children = append(children, cchildren...)
	}
	return children
}

func (store *BlockStore) FindChildren(parentID string) []string {
	store.RLock()
	defer store.RUnlock()
	return store.findChildren(parentID)
}

func (store *BlockStore) findChildren(parentID string) []string {
	children := []string{}
	for _, block := range store.blocks {
		for _, parent := range block.Event.Parents {
			if parent == parentID {
				children = append(children, block.Hash)
			}
		}
	}
	return children
}

func (store *BlockStore) FindTip(graphRootID string) ([]string, error) {
	store.RLock()
	defer store.RUnlock()
	return store.findTip(graphRootID)
}

func (store *BlockStore) findTip(graphRootID string) ([]string, error) {
	children := store.findChildren(graphRootID)
	if len(children) == 0 {
		return []string{graphRootID}, nil
	}
	tips := []string{}
	for _, childID := range children {
		ctips, _ := store.findTip(childID)
		if len(ctips) > 0 {
			tips = append(tips, ctips...)
		}
	}
	return SliceUniq(tips), nil
}

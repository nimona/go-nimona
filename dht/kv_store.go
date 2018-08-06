package dht

import (
	"sync"

	"github.com/nimona/go-nimona/blocks"
)

type Store struct {
	// TODO replace with async maps
	values    sync.Map
	providers sync.Map
	lock      sync.RWMutex
}

func newStore() (*Store, error) {
	s := &Store{
		values:    sync.Map{},
		providers: sync.Map{},
		lock:      sync.RWMutex{},
	}
	return s, nil
}

func (s *Store) PutProvider(block *blocks.Block) error {
	// TODO verify payload type
	s.providers.Store(block.ID(), block)
	return nil
}

func (s *Store) GetProviders(blockID string) ([]*blocks.Block, error) {
	bls := []*blocks.Block{}
	s.providers.Range(func(k, v interface{}) bool {
		block := v.(*blocks.Block)
		payload := block.Payload.(Provider)
		for _, id := range payload.BlockIDs {
			if id == blockID {
				bls = append(bls, block)
				break
			}
		}
		return true
	})

	return bls, nil
}

// GetAllProviders returns all providers and the values they are providing
func (s *Store) GetAllProviders() ([]*blocks.Block, error) {
	bls := []*blocks.Block{}
	s.providers.Range(func(k, v interface{}) bool {
		block := v.(*blocks.Block)
		bls = append(bls, block)
		return true
	})

	return bls, nil
}

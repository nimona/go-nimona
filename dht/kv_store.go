package dht

import (
	"sync"

	"github.com/nimona/go-nimona/net"
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

func (s *Store) PutProvider(block *net.Block) error {
	// TODO verify payload type
	s.providers.Store(block.Metadata.ID, block)
	return nil
}

func (s *Store) GetProviders(blockID string) ([]*net.Block, error) {
	blocks := []*net.Block{}
	s.providers.Range(func(k, v interface{}) bool {
		block := v.(*net.Block)
		payload := block.Payload.(PayloadProvider)
		for _, id := range payload.BlockIDs {
			if id == blockID {
				blocks = append(blocks, block)
				break
			}
		}
		return true
	})

	return blocks, nil
}

func (s *Store) PutValue(block *net.Block) error {
	// TODO verify payload type
	s.values.Store(block.Metadata.ID, block)
	return nil
}

func (s *Store) GetValue(key string) (string, error) {
	v, ok := s.values.Load(key)
	if !ok {
		return "", ErrNotFound
	}

	return v.(string), nil
}

// GetAllProviders returns all providers and the values they are providing
func (s *Store) GetAllProviders() ([]*net.Block, error) {
	blocks := []*net.Block{}
	s.providers.Range(func(k, v interface{}) bool {
		block := v.(*net.Block)
		blocks = append(blocks, block)
		return true
	})

	return blocks, nil
}

// GetAllValues returns all the key value pairs we know about
func (s *Store) GetAllValues() ([]*net.Block, error) {
	blocks := []*net.Block{}
	s.values.Range(func(k, v interface{}) bool {
		block := v.(*net.Block)
		blocks = append(blocks, block)
		return true
	})

	return blocks, nil
}

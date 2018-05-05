package blx

import "encoding/json"

type memoryStore struct {
	data map[string][]byte
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		data: map[string][]byte{},
	}
}
func (m *memoryStore) Store(key string, bl *Block) error {
	d, err := json.Marshal(*bl)
	if err != nil {
		return err
	}
	m.data[key] = d

	return nil
}

func (m *memoryStore) Get(key string) (*Block, error) {
	if bl, ok := m.data[key]; ok {
		block := Block{}
		err := json.Unmarshal(bl, &block)
		if err != nil {
			return nil, err
		}

		return &block, nil
	}
	return nil, ErrNotFound
}

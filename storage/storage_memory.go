package storage

// import "nimona.io/go/blocks"

// type memoryStore struct {
// 	data map[string]*blocks.Block
// }

// // NewMemoryStore creates a new in memory store
// func NewMemoryStore() Storage {
// 	return &memoryStore{
// 		data: map[string]*blocks.Block{},
// 	}
// }
// func (m *memoryStore) Store(key string, block *blocks.Block) error {
// 	m.data[key] = block

// 	return nil
// }

// func (m *memoryStore) Get(key string) (*blocks.Block, error) {
// 	block, ok := m.data[key]
// 	if !ok {
// 		return nil, ErrNotFound
// 	}

// 	return block, nil
// }

// func (m *memoryStore) List() ([]string, error) {
// 	results := make([]string, 0, 0)
// 	for k, _ := range m.data {
// 		results = append(results, k)
// 	}

// 	if len(results) == 0 {
// 		return []string{}, ErrEmpty
// 	}
// 	return results, nil
// }

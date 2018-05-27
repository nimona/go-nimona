package blx

type memoryStore struct {
	data map[string]*Block
}

func NewMemoryStore() *memoryStore {
	return &memoryStore{
		data: map[string]*Block{},
	}
}
func (m *memoryStore) Store(key string, block *Block) error {
	m.data[key] = block

	return nil
}

func (m *memoryStore) Get(key string) (*Block, error) {
	block, ok := m.data[key]
	if !ok {
		return nil, ErrNotFound
	}

	return block, nil
}

func (m *memoryStore) List() ([]*string, error) {
	results := make([]*string, 0, 0)
	for k, _ := range m.data {
		results = append(results, &k)
	}

	if len(results) == 0 {
		return []*string{}, ErrEmpty
	}
	return results, nil
}

package kv

// Store ...
type Store interface {
	Store(key string, value []byte) error
	Get(key string) ([]byte, error)
	List() ([]string, error)
	Scan(prefix string) ([]string, error)
}

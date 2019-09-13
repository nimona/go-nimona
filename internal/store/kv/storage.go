package kv

// Store ...
type Store interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, error)
	Check(key string) error
	Remove(key string) error
	List() ([]string, error)
	Scan(prefix string) ([]string, error)
}

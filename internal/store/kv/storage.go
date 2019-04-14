package kv

// Store ...
type Store interface {
	Store(string, []byte) error
	Get(string) ([]byte, error)
	List() ([]string, error)
}

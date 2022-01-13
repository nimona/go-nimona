package configstore

type (
	Store interface {
		Put(key string, value string) error
		Get(key string) (string, error)
		Remove(key string) error
		List() (map[string]string, error)
	}
)

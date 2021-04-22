package preferences

type (
	Preferences interface {
		Put(key string, value string) error
		Get(key string) (string, error)
		List() (map[string]string, error)
	}
)

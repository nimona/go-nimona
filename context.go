package fabric

type contextKey string

func (c contextKey) String() string {
	return "nimona:fabric:" + string(c)
}

package fabric

type TCP struct {
}

func (t *TCP) Dial(addr string) (Conn, error) {
	return nil, nil
}

package fabric

type Transport interface{
	Dial(addr string) (Conn, error)
}

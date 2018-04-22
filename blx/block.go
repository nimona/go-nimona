package blx

type Block struct {
	Key  string
	Data []byte
	Meta map[string][]byte
}

package blocks

type Signer interface {
	Sign([]byte) ([]byte, error)
}

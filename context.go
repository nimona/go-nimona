package nimona

type RequestContext struct {
	Identity      *Identity
	PublicKey     PublicKey
	PrivateKey    PrivateKey
	DocumentStore *DocumentStore
}

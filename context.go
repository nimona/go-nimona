package nimona

type RequestContext struct {
	Identity   *Identity
	PublicKey  PublicKey
	PrivateKey PrivateKey
}

func NewRequestContext(identity *Identity, sk PrivateKey, pk PublicKey) *RequestContext {
	return &RequestContext{
		Identity:   identity,
		PrivateKey: sk,
		PublicKey:  pk,
	}
}

package nimona

// TODO(geoah) refactor to use IdentityStore
type RequestContext struct {
	KeyGraphID    KeyGraphID
	PublicKey     PublicKey
	PrivateKey    PrivateKey
	DocumentStore *DocumentStore
}

type SigningContext struct {
	KeyGraphID KeyGraphID
	PrivateKey PrivateKey
}

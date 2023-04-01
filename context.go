package nimona

// TODO(geoah) refactor to use IdentityStore
type RequestContext struct {
	Identity      *Identity
	PublicKey     PublicKey
	PrivateKey    PrivateKey
	DocumentStore *DocumentStore
}

type SigningContext struct {
	Identity   *Identity
	PrivateKey PrivateKey
}

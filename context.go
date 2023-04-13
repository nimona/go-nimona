package nimona

// TODO(geoah) refactor to use IdentityStore
type RequestContext struct {
	KeygraphID    KeygraphID
	PublicKey     PublicKey
	PrivateKey    PrivateKey
	DocumentStore *DocumentStore
}

type SigningContext struct {
	KeygraphID KeygraphID
	PrivateKey PrivateKey
}

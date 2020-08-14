package keychain

type KeyType int

const (
	// PrimaryPeerKey defines the key pair the peer identifies itself with.
	// There can only be one.
	PrimaryPeerKey KeyType = 0
	// PrimaryIdentityKey defines an identity key that has signed the peer's
	// keys.
	PrimaryIdentityKey KeyType = 1
)

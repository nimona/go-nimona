package keychain

type KeyType int

const (
	// PrimaryPeerKey defines the key pair the peer identifies itself with.
	// There can only be one.
	PrimaryPeerKey KeyType = 0
	// PeerKey defines the additional keys for this peer.
	PeerKey KeyType = 1
	// IdentityKey defines an identity key that has signed the peer's keys.
	IdentityKey KeyType = 9
)

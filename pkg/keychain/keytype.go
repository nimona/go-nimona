package keychain

type keytype int

const (
	// PrimaryPeerKey defines the key pair the peer identifies itself with.
	// There can only be one.
	PrimaryPeerKey keytype = 0
	// PeerKey defines the additional keys for this peer.
	PeerKey keytype = 1
	// IdentityKey defines an identity key that has signed the peer's keys.
	IdentityKey keytype = 9
)

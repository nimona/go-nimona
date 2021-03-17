package crypto

type (
	KeyType   uint64
	PublicKey interface {
		// Equals(PublicKey) bool
		// Raw() ([]byte, error)
		Type() KeyType
		String() string
		MarshalString() (string, error)
		UnmarshalString(string) error
	}
	PrivateKey interface {
		PublicKey
		PublicKey() PublicKey
	}
)

const (
	cidEd25519Private = 0x1300 // well known value
	cidEd25519Public  = 0xED   // well known value

	PeerKey     KeyType = 0x6E00 // codec code for nimona _peer_ key
	IdentityKey KeyType = 0x6E01 // codec code for nimona _identity_ key
)

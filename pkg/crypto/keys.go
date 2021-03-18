package crypto

type (
	KeyType      uint64
	KeyAlgorithm uint64
	// PublicKey interface {
	// 	// Equals(PublicKey) bool
	// 	// Raw() ([]byte, error)
	// 	Type() KeyType
	// 	String() string
	// 	MarshalString() (string, error)
	// 	UnmarshalString(string) error
	// }
	// PrivateKey interface {
	// 	PublicKey
	// 	PublicKey() PublicKey
	// }
)

const (
	Ed25519Private KeyAlgorithm = 0x1300 // well known value
	Ed25519Public  KeyAlgorithm = 0xED   // well known value

	PeerKey     KeyType = 0x6E00 // codec code for nimona _peer_ key
	IdentityKey KeyType = 0x6E01 // codec code for nimona _identity_ key
)

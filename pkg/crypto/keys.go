package crypto

type (
	KeyUsage     uint64
	KeyAlgorithm uint64
)

const (
	Ed25519Private KeyAlgorithm = 0x1300 // well known value
	Ed25519Public  KeyAlgorithm = 0xED   // well known value

	PeerKey     KeyUsage = 0x6E00 // codec code for nimona _peer_ key
	IdentityKey KeyUsage = 0x6E01 // codec code for nimona _identity_ key
)

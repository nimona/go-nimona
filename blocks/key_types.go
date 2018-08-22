package blocks

// Supported values for KeyType
const (
	EC             = "EC"  // Elliptic Curve
	InvalidKeyType = ""    // Invalid KeyType
	OctetSeq       = "oct" // Octet sequence (used to represent symmetric keys)
	RSA            = "RSA" // RSA
)

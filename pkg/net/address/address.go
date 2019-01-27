package address

// Address to allow net and others to dial peers, identities, etc.
type Address struct {
	Type    string
	Address string
}

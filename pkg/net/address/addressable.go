package address

// Addressable is an interface to allow net and other packages to accept
// anything that has an address
type Addressable interface {
	Address() Address
}

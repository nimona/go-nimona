package address

type (
	Addressable interface {
		Addresses() Addresses
	}
	Addresses []string
)

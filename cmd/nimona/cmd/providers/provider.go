package providers

// Provider is any system that nimona can be deployed, remote or local
type Provider interface {
	// NewInstance creates a new server
	NewInstance(name string) error
}

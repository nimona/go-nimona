package providers

import "errors"

var (
	// ErrNoToken returned when no token has beed provided
	ErrNoToken = errors.New("missing token")
)

// Provider is any system that nimona can be deployed, remote or local
type Provider interface {
	// NewInstance creates a new server
	NewInstance(name, sshFingerprint, size, region string) (ip string,
		err error)
}

package fabric

import (
	"errors"
)

var (
	// ErrNoTransport for when there is no transport with which to dial the address
	ErrNoTransport = errors.New("Could not dial with available transports")
	// ErrInvalidProtocol when our handler doesn't know about a protocol in the
	ErrInvalidProtocol = errors.New("No such protocol")
	// errNoMoreProtocols when fabric cannot deal with any more
	errNoMoreProtocols = errors.New("No more protocols")
)

// New instance of fabric
func New(protocols ...Protocol) *Fabric {
	baseAddress := make([]string, len(protocols))
	for i, protocol := range protocols {
		baseAddress[i] = protocol.Name()
	}
	f := &Fabric{
		base:       baseAddress,
		transports: []Transport{},
		protocols:  map[string]Protocol{},
	}
	for _, m := range protocols {
		f.AddProtocol(m)
	}
	return f
}

// Fabric manages transports and protocols, and deals with Dialing.
type Fabric struct {
	base       []string
	transports []Transport
	protocols  map[string]Protocol
}

// AddTransport for dialing to the outside world
func (f *Fabric) AddTransport(tr Transport) error {
	f.transports = append(f.transports, tr)
	return nil
}

// AddProtocol for both client and server
func (f *Fabric) AddProtocol(protocol Protocol) error {
	f.protocols[protocol.Name()] = protocol
	return nil
}

// GetAddresses returns a list of addresses for all the current transports
func (f *Fabric) GetAddresses() []string {
	addresses := []string{}
	for _, tr := range f.transports {
		addresses = append(addresses, tr.Addresses()...)
	}

	return addresses
}

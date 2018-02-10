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

var (
	// ContextKeyRequestID attached to each request
	ContextKeyRequestID = contextKey("request_id")
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

// // AddHandler for server
// func (f *Fabric) AddHandler(m Handler) error {
// 	return f.AddHandlerFunc(m.Name(), m.Handle)
// }

// // AddNegotiator for client
// func (f *Fabric) AddNegotiator(m Negotiator) error {
// 	return f.AddNegotiatorFunc(m.Name(), m.Negotiate)
// }

// // AddHandlerFunc for server
// func (f *Fabric) AddHandlerFunc(r string, h HandlerFunc) error {
// 	f.handlers[r] = h
// 	return nil
// }

// // AddNegotiatorFunc for client
// func (f *Fabric) AddNegotiatorFunc(n string, ng NegotiatorFunc) error {
// 	f.negotiators[n] = ng
// 	return nil
// }

// GetAddresses returns a list of addresses for all the current transports
func (f *Fabric) GetAddresses() []string {
	addresses := []string{}
	for _, tr := range f.transports {
		addresses = append(addresses, tr.Addresses()...)
	}

	return addresses
}

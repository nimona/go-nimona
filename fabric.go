package fabric

import (
	"context"
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
func New(ctx context.Context) *Fabric {
	f := &Fabric{
		context:    ctx,
		transports: []*transportWithProtocols{},
		protocols:  map[string]Protocol{},
	}
	return f
}

// Fabric manages transports and protocols, and deals with Dialing.
type Fabric struct {
	context    context.Context
	transports []*transportWithProtocols
	protocols  map[string]Protocol
}

type transportWithProtocols struct {
	Transport Transport
	Handler   HandlerFunc
	// Negotiator    NegotiatorFunc
	BaseProtocols []string
}

// AddTransport for dialing to the outside world
func (f *Fabric) AddTransport(transport Transport, protocols []Protocol) error {
	protocolNames := []string{}
	for _, pr := range protocols {
		protocolNames = append(protocolNames, pr.Name())
	}
	hWrapper := NewTransportWrapper(protocolNames)
	hchain := append([]Protocol{hWrapper}, protocols...)
	tr := &transportWithProtocols{
		Transport: transport,
		Handler:   HandlerChain(hchain...),
		// Negotiator:    negotiatorChain(protocols...),
		BaseProtocols: []string{},
	}
	f.transports = append(f.transports, tr)
	return transport.Listen(f.context, tr.Handler)
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
		addresses = append(addresses, tr.Transport.Addresses()...)
	}

	return addresses
}

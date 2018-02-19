package fabric

import (
	"context"
	"errors"

	protocol "github.com/nimona/go-nimona-fabric/protocol"
	transport "github.com/nimona/go-nimona-fabric/transport"
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
		protocols:  map[string]protocol.Protocol{},
	}
	return f
}

// Fabric manages transports and protocols, and deals with Dialing.
type Fabric struct {
	context    context.Context
	transports []*transportWithProtocols
	protocols  map[string]protocol.Protocol
}

type transportWithProtocols struct {
	Transport transport.Transport
	Handler   protocol.HandlerFunc
	// Negotiator    NegotiatorFunc
	BaseProtocols []string
}

// AddTransport for dialing to the outside world
func (f *Fabric) AddTransport(transport transport.Transport, protocols []protocol.Protocol) error {
	protocolNames := []string{}
	for _, pr := range protocols {
		protocolNames = append(protocolNames, pr.Name())
	}
	hWrapper := protocol.NewTransportWrapper(protocolNames)
	hchain := append([]protocol.Protocol{hWrapper}, protocols...)
	tr := &transportWithProtocols{
		Transport: transport,
		Handler:   protocol.HandlerChain(hchain...),
		// Negotiator:    negotiatorChain(protocols...),
		BaseProtocols: []string{},
	}
	f.transports = append(f.transports, tr)
	return transport.Listen(f.context, tr.Handler)
}

// AddProtocol for both client and server
func (f *Fabric) AddProtocol(protocol protocol.Protocol) error {
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

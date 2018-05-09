package net

import (
	"context"
	"errors"
	"strings"
	"sync"
)

var (
	// ErrInvalidProtocol when our handler doesn't know about a protocol in the
	ErrInvalidProtocol = errors.New("No such protocol")
	// errNoMoreProtocols when net cannot deal with any more protocols
	errNoMoreProtocols = errors.New("No more protocols")
)

// Net is our network interface for net
type Net interface {
	DialContext(ctx context.Context, address string) (context.Context, Conn, error)
	AddTransport(transport Transport, protocols ...Protocol) error
	AddProtocols(protocols ...Protocol) error
	GetAddresses() []string
	GetProtocols() map[string][]string
}

// New instance of net
func New(ctx context.Context) Net {
	f := &nnet{
		context:    ctx,
		transports: []*transportWithProtocols{},
		protocols:  map[string]Protocol{},
		lock:       &sync.Mutex{},
	}
	return f
}

// Net manages transports and protocols, and deals with Dialing.
type nnet struct {
	context    context.Context
	transports []*transportWithProtocols
	protocols  map[string]Protocol
	lock       *sync.Mutex
}

type transportWithProtocols struct {
	Transport     Transport
	Handler       HandlerFunc
	BaseProtocols []string
}

// AddTransport for dialing to the outside world
func (f *nnet) AddTransport(transport Transport, protocols ...Protocol) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	protocolNames := []string{}
	for _, pr := range protocols {
		protocolNames = append(protocolNames, pr.Name())
		if err := f.addProtocols(pr); err != nil {
			return err
		}
	}
	hWrapper := NewTransportWrapper(protocolNames)
	hchain := append([]Protocol{hWrapper}, protocols...)
	tr := &transportWithProtocols{
		Transport:     transport,
		Handler:       HandlerChain(hchain...),
		BaseProtocols: protocolNames,
	}
	f.transports = append(f.transports, tr)
	return transport.Listen(f.context, tr.Handler)
}

// AddProtocols for both client and server
func (f *nnet) AddProtocols(protocols ...Protocol) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.addProtocols(protocols...)
}

// addProtocols without locking for internal use
func (f *nnet) addProtocols(protocols ...Protocol) error {
	for _, protocol := range protocols {
		f.protocols[protocol.Name()] = protocol
	}
	return nil
}

// GetAddresses returns a list of addresses for all the current transports
func (f *nnet) GetAddresses() []string {
	// TODO use different lock
	f.lock.Lock()
	defer f.lock.Unlock()
	addresses := []string{}
	for _, tr := range f.transports {
		addresses = append(addresses, tr.Transport.GetAddresses()...)
	}
	return addresses
}

// GetAddresses returns a list of addresses for all the current transports
func (f *nnet) GetProtocols() map[string][]string {
	// TODO use different lock
	f.lock.Lock()
	defer f.lock.Unlock()
	protocols := map[string][]string{}
	for _, tr := range f.transports {
		// go through all tranports
		for _, transportAddress := range tr.Transport.GetAddresses() {
			// check if they have any protocols
			if len(tr.BaseProtocols) == 0 {
				continue
			}
			// gather addresses from all protocols
			protocolAddresses := []string{}
			for _, protocolName := range tr.BaseProtocols {
				// add all protocol addresses
				protocol := f.protocols[protocolName]
				protocolAddresses = joinStringMatrix(protocolAddresses, protocol.GetAddresses())
			}
			// find the last part
			for _, protocolAddress := range protocolAddresses {
				parts := strings.Split(protocolAddress, "/")
				lastPart := parts[len(parts)-1]
				if _, ok := protocols[lastPart]; !ok {
					protocols[lastPart] = []string{}
				}
				fullAddress := transportAddress + "/" + protocolAddress
				protocols[lastPart] = append(protocols[lastPart], fullAddress)
			}

		}
	}
	return protocols
}

func joinStringMatrix(a, b []string) []string {
	r := []string{}
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	for _, ia := range a {
		for _, ib := range b {
			r = append(r, strings.Join([]string{ia, ib}, "/"))
		}
	}
	return r
}

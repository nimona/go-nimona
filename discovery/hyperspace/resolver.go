package hyperspace

import (
	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	"nimona.io/go/net"
	"nimona.io/go/peers"
)

var (
	typePeerInfoRequest = peers.PeerInfoRequest{}.GetType()
	typePeerInfo        = peers.PeerInfo{}.GetType()
)

type Resolver struct {
	store *Store
}

func NewResolver(key *crypto.Key, network net.Network, exchange net.Exchange,
	bootstrapAddresses []string) (*Resolver, error) {
	return &Resolver{
		store: NewStore(),
	}, nil
}

func (r *Resolver) Resolve(q *peers.PeerInfoRequest) (*peers.PeerInfo, error) {
	return nil, nil
}

func (r *Resolver) handleBlock(o *encoding.Object) error {
	switch o.GetType() {
	case typePeerInfoRequest:
		v := &peers.PeerInfoRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfoRequest(v)
	case typePeerInfo:
		v := &peers.PeerInfo{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfo(v)
	default:
		return nil
	}
	return nil
}

func (r *Resolver) handlePeerInfo(p *peers.PeerInfo) {
	// ...
}

func (r *Resolver) handlePeerInfoRequest(q *peers.PeerInfoRequest) {
	// ...
}

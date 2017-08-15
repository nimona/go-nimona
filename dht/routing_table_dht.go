package dht

import (
	"sync"

	net "github.com/nimona/go-nimona-net"
	"github.com/sirupsen/logrus"
)

// RoutingTableSimple ...
type RoutingTableSimple struct {
	localPeer net.Peer
	mx        sync.RWMutex
	store     map[string]net.Peer
}

// Save ...
func (rt *RoutingTableSimple) Save(peer net.Peer) error {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	// If peer exists update address
	if _, ok := rt.store[peer.ID]; ok {
		op := rt.store[peer.ID]
		op.Addresses = peer.Addresses
	}
	rt.store[peer.ID] = peer

	return nil
}

// Remove ...
func (rt *RoutingTableSimple) Remove(peer net.Peer) error {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	if _, ok := rt.store[peer.ID]; !ok {
		return ErrPeerNotFound
	}
	delete(rt.store, peer.ID)

	return nil
}

// Get ...
func (rt *RoutingTableSimple) Get(string string) (net.Peer, error) {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	pr, ok := rt.store[string]
	if !ok {
		return net.Peer{}, ErrPeerNotFound
	}

	return pr, nil
}

func (rt *RoutingTableSimple) GetPeerIDs() ([]string, error) {
	rt.mx.Lock()
	defer rt.mx.Unlock()
	ids := make([]string, len(rt.store))
	i := 0
	for _, peer := range rt.store {
		ids[i] = peer.ID
		i++
	}
	return ids, nil
}

func NewSimpleRoutingTable(nnet net.Network, localPeer net.Peer) *RoutingTableSimple {
	rt := &RoutingTableSimple{
		localPeer: localPeer,
		store:     make(map[string]net.Peer),
	}
	// handle new network peers
	nnet.RegisterPeerHandler(func(np net.Peer) error {
		if len(np.Addresses) == 0 {
			return nil
		}

		if np.ID == localPeer.ID {
			localPeer.Addresses = []string{}
			for _, addr := range np.Addresses {
				localPeer.Addresses = append(localPeer.Addresses, addr)
			}
		}

		logrus.WithField("np", np).Debugf("Handling incoming peer")
		if err := rt.Save(np); err != nil {
			logrus.WithError(err).Debugf("Could not add incoming peer")
			return err
		}
		logrus.Debugf("Saved incoming peer")
		return nil
	})

	return rt
}

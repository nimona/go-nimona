package dht

import (
	"sort"
	"sync"

	net "github.com/nimona/go-nimona-net"
	"github.com/sirupsen/logrus"
)

// RoutingTable ...
type RoutingTable struct {
	localPeer net.Peer
	mx        sync.RWMutex
	store     map[string]net.Peer
}

// Put upserts a peer
func (rt *RoutingTable) Put(peer net.Peer) error {
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

// Remove a peer
func (rt *RoutingTable) Remove(peer net.Peer) error {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	if _, ok := rt.store[peer.ID]; !ok {
		return ErrPeerNotFound
	}
	delete(rt.store, peer.ID)

	return nil
}

// Get a peer
func (rt *RoutingTable) Get(string string) (net.Peer, error) {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	pr, ok := rt.store[string]
	if !ok {
		return net.Peer{}, ErrPeerNotFound
	}

	return pr, nil
}

func (rt *RoutingTable) getPeerIDs() ([]string, error) {
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

// FindPeersNear accepts an ID and n and finds the n closest nodes to this id
// in the routing table
func (rt *RoutingTable) FindPeersNear(id string, n int) ([]net.Peer, error) {
	peers := []net.Peer{}

	ids, err := rt.getPeerIDs()
	if err != nil {
		logrus.WithError(err).Error("Failed to get peer ids")
		return peers, err
	}

	// slice to hold the distances
	dists := []distEntry{}
	for _, pid := range ids {
		entry := distEntry{
			id:   pid,
			dist: xor([]byte(id), []byte(pid)),
		}
		dists = append(dists, entry)
	}
	// Sort the distances
	sort.Slice(dists, func(i, j int) bool {
		return lessIntArr(dists[i].dist, dists[j].dist)
	})

	if n > len(dists) {
		n = len(dists)
	}
	// Append n the first n number of peers from the ids
	for _, de := range dists {
		p, err := rt.Get(de.id)
		if err != nil {
			logrus.WithError(err).WithField("ID", de.id).Error("Peer not found")
		}
		if len(p.Addresses) == 0 {
			continue
		}
		peers = append(peers, p)
		n--
		if n == 0 {
			break
		}
	}
	return peers, nil
}

func NewRoutingTable(nnet net.Network, localPeer net.Peer) *RoutingTable {
	rt := &RoutingTable{
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
		if err := rt.Put(np); err != nil {
			logrus.WithError(err).Debugf("Could not add incoming peer")
			return err
		}
		logrus.Debugf("Saved incoming peer")
		return nil
	})

	return rt
}

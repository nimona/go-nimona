package dht

import (
	"context"
	"fmt"

	net "github.com/nimona/go-nimona-net"
	"github.com/sirupsen/logrus"
)

const numPeersNear int = 3

type query struct {
	id             string
	nonce          string
	net            net.Network
	closestPeer    net.Peer
	shortlistPeers []net.Peer
	routingTable   RoutingTable
	incoming       chan string
	results        chan net.Peer
	peersQueue     chan net.Peer
	dht            *DHTNode
}

func (q *query) Run(ctx context.Context) (chan net.Peer, error) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case p := <-q.peersQueue: // listen for incoming internal queries
				msg := &Message{
					OriginPeer:  q.dht.GetLocalPeer(),
					Nonce:       q.nonce,
					QueryPeerID: q.id,
				}
				q.dht.sendMsgPeer(MESSAGE_TYPE_FIND_NODE, msg, p.ID)
				// q.dht.queryPeer(ctx, p, q.incomming)

			case v := <-q.incoming:
				logrus.Info("Incoming: ", v)
				// 	if strings.Contains(v, "peer/") {
				// 		p, err := peerFromValue(v)
				// 		q.peersQueue <- p
				// 	} else {
				// 		q.results <- v
				// 	}
			}
		}
	}()

	logrus.Debugf("Getting nearest peers")
	// Check peers in local store for distance
	// send message to the X closest peers
	lookupPeers, err := q.routingTable.FindPeersNear(q.id, numPeersNear)
	if err != nil {
		logrus.WithError(err).Error("Failed find peers near")
	}

	// Store the closest peer
	closestPeer := net.Peer{}
	if len(lookupPeers) > 0 {
		closestPeer = lookupPeers[0]
	}
	q.closestPeer = closestPeer

	logrus.WithField("peers", lookupPeers).Debugf("Asking nearest peers")
	for _, p := range lookupPeers {
		logrus.
			WithField("peer", p.ID).
			WithField("addr", p.Addresses).
			Infof("Asking peer for %s", q.id)
		q.peersQueue <- p
		// err := nd.s
	}

	return q.results, nil
}

func (q *query) queryPeer(ctx context.Context, peer net.Peer) {
	fmt.Println(peer)
	// results := make(chan net.Peer)endMsgPeer(MESSAGE_TYPE_FIND_NODE, msg, p.ID)
	// if err != nil {
	// 	logrus.WithError(err).WithField(
	// 		"peer",
	// 		p.ID,
	// 	).Error("Failed to send message to peer")
	// }
	// q.dht.queryPeer(ctx, peer, message, results)
}

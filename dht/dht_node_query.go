package dht

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	net "github.com/nimona/go-nimona-net"
)

const numPeersNear int = 3

type query struct {
	id                string
	peerID            string
	dht               *DHTNode
	closestPeerID     string
	contactedPeers    []string
	incomingResponses chan findNodeResponse
	results           chan net.Peer
}

func (q *query) Run(ctx context.Context) {
	logger := logrus.WithField("resp", q.id)

	go func() {
		// close channel once we are done
		// defer close(q.results)

		// check if we already know about the peer
		if peer, err := q.dht.routingtable.Get(q.peerID); err == nil {
			// if so, return it and stop
			logger.Infof("Peer existed in local store")
			q.results <- peer
			return
		}

		// wait for something to happen
		for {
			select {
			case resp := <-q.incomingResponses:
				logger.Infof("Processing incoming response")
				// keep whether we found our peer
				found := false
				// keep whether we got closer than before
				closer := false
				// go through the results
				for _, peer := range resp.Peers {
					// check if we found the node
					if peer.ID == q.peerID {
						logger.WithField("peerID", q.peerID).Infof("Found peer")
						q.closestPeerID = peer.ID
						q.results <- peer
						found = true
					}
					// add everything to our store
					q.dht.putPeer(peer)
					// check if we got closer than before
					if q.closestPeerID == "" {
						q.closestPeerID = peer.ID
					} else {
						if comparePeers(q.closestPeerID, peer.ID, q.peerID) == peer.ID {
							q.closestPeerID = peer.ID
							closer = true
						}
					}
				}
				// send next request
				if !found && closer {
					go q.next()
				}

			case <-time.After(time.Second * 5):
				logrus.Warnf("Time has passed")
				close(q.results)
				return

			case <-ctx.Done():
				logrus.Warnf("CTX was done")
				close(q.results)
				return
			}
		}
	}()

	// start looking for the node
	go q.next()
}

func (q *query) next() {
	// find closest peers
	cps, err := q.dht.routingtable.FindPeersNear(q.id, numPeersNear*5)
	if err != nil {
		logrus.WithError(err).Error("Failed find peers near")
		return
	}
	// create request
	req := findNodeRequest{
		QueryID:     q.id,
		OriginPeer:  q.dht.GetLocalPeer(),
		QueryPeerID: q.peerID,
	}
	// keep track of how many we've sent to
	sent := 0
	// go through closest peers
	for _, cp := range cps {
		// skip the ones we've already asked
		for _, pid := range q.contactedPeers {
			if pid == cp.ID {
				continue
			}
		}
		// ask peer
		logrus.WithField("peerID", cp.ID).WithField("querID", req.QueryID).Infof("Asking peer")
		q.dht.sendMsgPeer(MESSAGE_TYPE_FIND_NODE_REQ, req, cp.ID)
		// mark peer as contacted
		q.contactedPeers = append(q.contactedPeers, cp.ID)
		// stop once we reached the limit
		sent++
		if sent >= numPeersNear {
			return
		}
	}
}

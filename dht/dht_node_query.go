package dht

import (
	"context"
	"strings"
	"sync"
	"time"

	logrus "github.com/sirupsen/logrus"
)

const numPeersNear int = 15

type query struct {
	id               string
	key              string
	dht              *DHTNode
	closestPeerID    string
	contactedPeers   []string
	incomingMessages chan messagePut
	results          chan string
	lock             *sync.RWMutex
}

func (q *query) Run(ctx context.Context) {
	logger := logrus.WithField("resp", q.key)

	go func() {
		// close channel once we are done
		// defer close(q.results)

		// send what we know about the key
		if pair, err := q.dht.store.Get(q.key); err == nil {
			// if so, return it
			logger.Infof("Peer existed in local store")
			for _, v := range pair {
				q.results <- v
			}
		}

		// wait for something to happen
		for {
			select {
			case msg := <-q.incomingMessages:
				logger.Infof("Processing incoming message")
				// check if we found the node
				if msg.Key == q.key {
					logger.WithField("key", q.key).Infof("Found peer")
					// send results
					for _, v := range msg.Values {
						q.results <- v
					}
				}
				// if pair is peer information consider it a closest peer
				if strings.Contains(msg.Key, KeyPrefixPeer) {
					// check if we got closer than before
					if q.closestPeerID == "" {
						q.closestPeerID = msg.Key
					} else {
						if comparePeers(q.closestPeerID, msg.Key, q.key) == msg.Key {
							q.closestPeerID = msg.Key
							go q.next()
						}
					}
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
	q.lock.Lock()
	defer q.lock.Unlock()

	// find closest peers
	cps, err := q.dht.store.FindKeysNearestTo(KeyPrefixPeer, q.key, numPeersNear*10)
	if err != nil {
		logrus.WithError(err).Error("Failed find peers near")
		return
	}
	// create request
	req := messageGet{
		QueryID:    q.id,
		OriginPeer: q.dht.GetLocalPeer(),
		Key:        q.key,
	}
	// keep track of how many we've sent to
	sent := 0
	// go through closest peers
	for _, cp := range cps {
		// skip the ones we've already asked
		if in(cp, q.contactedPeers) {
			continue
		}
		// ask peer
		logrus.WithField("src", q.dht.GetLocalPeer().ID).WithField("dst", cp).WithField("queryKey", req.Key).WithField("id", q.id).WithField("cp", q.contactedPeers).Infof("Asking peer")
		q.dht.sendMessage(MessageTypeGet, req, trimKey(cp, KeyPrefixPeer))
		// mark peer as contacted
		q.contactedPeers = append(q.contactedPeers, cp)
		// stop once we reached the limit
		sent++
		if sent >= numPeersNear {
			return
		}
	}
}

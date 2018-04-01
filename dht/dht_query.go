package dht

import (
	"context"
	"sync"
	"time"

	"github.com/nimona/go-nimona/net"
	logrus "github.com/sirupsen/logrus"
)

const numPeersNear int = 15

type query struct {
	id               string
	key              string
	labels           map[string]string
	dht              *DHT
	closestPeerID    string
	contactedPeers   sync.Map
	incomingMessages chan messagePut
	results          chan net.Record
	// lock             *sync.RWMutex
}

func (q *query) Run(ctx context.Context) {
	logger := logrus.WithField("resp", q.key)

	go func() {
		// close channel once we are done
		// defer close(q.results)

		// send what we know about the key
		if pairs, err := q.dht.store.Filter(q.key, q.labels); err == nil {
			// if so, return it
			if len(pairs) > 0 {
				logger.Debug("Value existed in local store")
				for _, pair := range pairs {
					q.results <- pair
				}
			}
		}

		// wait for something to happen
		for {
			select {
			case msg := <-q.incomingMessages:
				logger.Debug("Processing incoming message")
				// check if we found the node
				persist := false
				pair := Pair{
					Key:    msg.Key,
					Value:  msg.Value,
					Labels: msg.Labels,
				}
				if msg.Key == q.key {
					logger.WithField("key", q.key).Debug("Found value")
					// persist the things we asked about
					persist = true
					// send results
					q.results <- pair
				}
				// store values we got
				q.dht.store.Put(msg.Key, msg.Value, msg.Labels, persist)
				// if pair is peer information consider it a closest peer
				if pair.GetLabel("protocol") == "peer" {
					// check if we got closer than before
					if q.closestPeerID == "" {
						q.closestPeerID = msg.Key
						go q.next()
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
	// find closest peers
	cps, err := q.dht.store.FindPeersNearestTo(q.key, numPeersNear)
	if err != nil {
		logrus.WithError(err).Error("Failed find peers near")
		return
	}

	// create request
	req := messageGet{
		QueryID:      q.id,
		OriginPeerID: q.dht.peerID,
		Key:          q.key,
		Labels:       q.labels,
	}
	// keep track of how many we've sent to
	sent := 0
	// go through closest peers
	for _, cp := range cps {
		// skip the ones we've already asked
		if _, ok := q.contactedPeers.Load(cp); ok {
			continue
		}
		// ask peer
		logrus.WithField("src", q.dht.peerID).
			WithField("dst", cp).
			WithField("queryKey", req.Key).
			WithField("id", q.id).
			WithField("cp", q.contactedPeers).
			Infof("Asking peer")
		q.dht.sendMessage(MessageTypeGet, req, cp)
		// mark peer as contacted
		q.contactedPeers.Store(cp, true)
		// stop once we reached the limit
		sent++
		if sent >= numPeersNear {
			return
		}
	}
}

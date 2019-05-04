package dht

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object"
)

const numPeersNear int = 15

type QueryType int

const (
	PeerInfoQuery QueryType = iota
	ProviderQuery
)

type query struct {
	dht              *DHT
	id               string
	key              string
	queryType        QueryType
	closestPeerID    string
	contactedPeers   sync.Map
	incomingPayloads chan interface{}
	outgoingPayloads chan interface{}
	logger           *zap.Logger
}

func (q *query) Run(ctx context.Context) {
	q.logger = log.Logger(ctx)
	go func() {
		// send what we know about the key
		switch q.queryType {
		case PeerInfoQuery:
			if peerInfo, err := q.dht.peerStore.Get(q.key); err == nil {
				if peerInfo == nil {
					q.logger.Warn("got nil peerInfo", zap.String("requestID", q.key))
					break
				}
				q.outgoingPayloads <- peerInfo
			}
		case ProviderQuery:
			if providers, err := q.dht.store.GetProviders(q.key); err == nil {
				for _, provider := range providers {
					q.outgoingPayloads <- provider
				}
			}
		}

		// and now, wait for something to happen
		for {
			select {
			case incPayload := <-q.incomingPayloads:
				switch payload := incPayload.(type) {
				case *peer.PeerInfo:
					q.outgoingPayloads <- payload
					// TODO next doesn't work
					// q.nextIfCloser(object.SenderPeerInfo.Metadata.Signer)
				case *Provider:
					// TODO check if id is in payload.ObjectIDs
					for _, objectID := range payload.ObjectIDs {
						if objectID == q.key {
							q.outgoingPayloads <- payload
							break
						}
					}
					// TODO next doesn't work
					// q.nextIfCloser(object.SenderPeerInfo.Metadata.Signer)
				}

			case <-time.After(maxQueryTime):
				close(q.outgoingPayloads)
				return

			case <-ctx.Done():
				close(q.outgoingPayloads)
				return
			}
		}
	}()

	// start looking for the node
	go q.next()
}

func (q *query) nextIfCloser(newPeerID string) {
	if q.closestPeerID == "" {
		q.closestPeerID = newPeerID
		q.next()
	} else {
		// find closest peer
		closestPeers, err := q.dht.FindPeersClosestTo(q.key, 1)
		if err != nil {
			// TODO log error
			return
		}
		if len(closestPeers) == 0 {
			return
		}
		closestPeerID := closestPeers[0].Fingerprint()
		if comparePeers(q.closestPeerID, closestPeerID, q.key) == closestPeerID {
			q.closestPeerID = closestPeerID
			q.next()
		}
	}
}

func (q *query) next() {
	// find closest peers
	closestPeers, err := q.dht.FindPeersClosestTo(q.key, numPeersNear)
	if err != nil {
		q.logger.Warn("Failed find peers near", zap.Error(err))
		return
	}

	peersToAsk := []*crypto.PublicKey{}
	for _, peerInfo := range closestPeers {
		// skip the ones we've already asked
		if _, ok := q.contactedPeers.Load(peerInfo.Fingerprint()); ok {
			continue
		}
		peersToAsk = append(peersToAsk, peerInfo.Signature.PublicKey)
		q.contactedPeers.Store(peerInfo.Fingerprint(), true)
	}

	signer := q.dht.key

	var o *object.Object
	switch q.queryType {
	case PeerInfoQuery:
		req := &PeerInfoRequest{
			RequestID: q.id,
			PeerID:    q.key,
		}
		o = req.ToObject()
		err = crypto.Sign(o, signer)
	case ProviderQuery:
		req := &ProviderRequest{
			RequestID: q.id,
			Key:       q.key,
		}
		o = req.ToObject()
		err = crypto.Sign(o, signer)
	default:
		return
	}

	if err != nil {
		return
	}

	ctx := context.Background()
	logger := log.Logger(ctx)
	for _, peer := range peersToAsk {
		addr := "peer:" + peer.Fingerprint()
		if err := q.dht.exchange.Send(ctx, o, addr); err != nil {
			logger.Warn("query next could not send", zap.Error(err), zap.String("peerID", peer.Fingerprint()))
		}
	}
}

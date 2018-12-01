package dht

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	"nimona.io/go/log"
	"nimona.io/go/peers"
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
			if peerInfo, err := q.dht.addressBook.GetPeerInfo(q.key); err == nil {
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
				case *peers.PeerInfo:
					q.outgoingPayloads <- payload
					// TODO next doesn't work
					// q.nextIfCloser(block.SenderPeerInfo.Metadata.Signer)
				case *Provider:
					// TODO check if id is in payload.BlockIDs
					for _, blockID := range payload.BlockIDs {
						if blockID == q.key {
							q.outgoingPayloads <- payload
							break
						}
					}
					// TODO next doesn't work
					// q.nextIfCloser(block.SenderPeerInfo.Metadata.Signer)
				}

			// case <-time.After(maxQueryTime):
			// 	close(q.outgoingPayloads)
			// 	return

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
		closestPeerID := closestPeers[0].Thumbprint()
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

	peersToAsk := []*crypto.Key{}
	for _, peerInfo := range closestPeers {
		// skip the ones we've already asked
		if _, ok := q.contactedPeers.Load(peerInfo.Thumbprint()); ok {
			continue
		}
		peersToAsk = append(peersToAsk, peerInfo.SignerKey)
		q.contactedPeers.Store(peerInfo.Thumbprint(), true)
	}

	signer := q.dht.addressBook.GetLocalPeerKey()

	var o *encoding.Object
	switch q.queryType {
	case PeerInfoQuery:
		req := &PeerInfoRequest{
			RequestID: q.id,
			PeerID:    q.key,
		}
		o, err = crypto.Sign(req.ToObject(), signer)
	case ProviderQuery:
		req := &ProviderRequest{
			RequestID: q.id,
			Key:       q.key,
		}
		o, err = crypto.Sign(req.ToObject(), signer)
	default:
		return
	}

	if err != nil {
		return
	}

	ctx := context.Background()
	logger := log.Logger(ctx)
	for _, peer := range peersToAsk {
		addr := "peer:" + peer.HashBase58()
		if err := q.dht.exchange.Send(ctx, o, addr); err != nil {
			logger.Debug("query next could not send", zap.Error(err), zap.String("peerID", peer.HashBase58()))
		}
	}
}

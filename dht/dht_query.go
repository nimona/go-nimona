package dht

import (
	"context"
	"sync"
	"time"

	"github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/log"
	"github.com/nimona/go-nimona/peers"
	logrus "github.com/sirupsen/logrus"
	"go.uber.org/zap"
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
}

func (q *query) Run(ctx context.Context) {
	logger := log.Logger(ctx)
	go func() {
		// send what we know about the key
		switch q.queryType {
		case PeerInfoQuery:
			if peerInfo, err := q.dht.addressBook.GetPeerInfo(q.key); err == nil {
				if peerInfo == nil {
					logger.Warn("got nil peerInfo", zap.String("requestID", q.key))
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
					// q.nextIfCloser(block.SenderPeerInfo.Metadata.Signer)
				case *Provider:
					// TODO check if id is in payload.BlockIDs
					for _, blockID := range payload.BlockIDs {
						if blockID == q.key {
							q.outgoingPayloads <- payload
							break
						}
					}
					// q.nextIfCloser(block.SenderPeerInfo.Metadata.Signer)
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
		logrus.WithError(err).Error("Failed find peers near")
		return
	}

	peersToAsk := []*blocks.Key{}
	for _, peerInfo := range closestPeers {
		// skip the ones we've already asked
		if _, ok := q.contactedPeers.Load(peerInfo.Thumbprint()); ok {
			continue
		}
		peersToAsk = append(peersToAsk, peerInfo.Signature.Key)
		q.contactedPeers.Store(peerInfo.Thumbprint(), true)
	}

	// var payloadType string
	var req interface{}

	switch q.queryType {
	case PeerInfoQuery:
		// payloadType = PeerInfoRequestType
		req = PeerInfoRequest{
			RequestID: q.id,
			PeerID:    q.key,
		}
	case ProviderQuery:
		// payloadType = ProviderRequestType
		req = ProviderRequest{
			RequestID: q.id,
			Key:       q.key,
		}
	default:
		return
	}

	ctx := context.Background()
	signer := q.dht.addressBook.GetLocalPeerInfo().Key
	// block := blocks.NewEphemeralBlock(payloadType, req)
	// block.SetHeader("requestID", q.id)
	for _, peer := range peersToAsk {
		if err := q.dht.exchange.Send(ctx, req, peer, blocks.SignWith(signer)); err != nil {
			panic(err)
		}
	}
}

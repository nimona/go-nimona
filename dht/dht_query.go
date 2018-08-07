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
	dht            *DHT
	id             string
	key            string
	queryType      QueryType
	closestPeerID  string
	contactedPeers sync.Map
	incomingBlocks chan *blocks.Block
	outgoingBlocks chan *blocks.Block
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
				q.outgoingBlocks <- peerInfo.Block
			}
		case ProviderQuery:
			if providers, err := q.dht.store.GetProviders(q.key); err == nil {
				for _, provider := range providers {
					if provider == nil {
						logger.Warn("got nil provider", zap.String("requestID", q.key))
						continue
					}
					q.outgoingBlocks <- provider
				}
			}
		}

		// and now, wait for something to happen
		for {
			select {
			case incBlock := <-q.incomingBlocks:
				incPayload := incBlock.Payload
				switch payload := incPayload.(type) {
				case peers.PeerInfo:
					q.outgoingBlocks <- incBlock
					// q.nextIfCloser(block.SenderPeerInfo.Metadata.Signer)
				case Provider:
					// TODO check if id is in payload.BlockIDs
					for _, blockID := range payload.BlockIDs {
						if blockID == q.key {
							q.outgoingBlocks <- incBlock
							break
						}
					}
					// q.nextIfCloser(block.SenderPeerInfo.Metadata.Signer)
				}

			case <-time.After(maxQueryTime):
				close(q.outgoingBlocks)
				return

			case <-ctx.Done():
				close(q.outgoingBlocks)
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
		closestPeerID := closestPeers[0].ID
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

	peersToAsk := []string{}
	for _, peerInfo := range closestPeers {
		// skip the ones we've already asked
		if _, ok := q.contactedPeers.Load(peerInfo.ID); ok {
			continue
		}
		peersToAsk = append(peersToAsk, peerInfo.ID)
		q.contactedPeers.Store(peerInfo.ID, true)
	}

	var payloadType string
	var req interface{}

	switch q.queryType {
	case PeerInfoQuery:
		payloadType = PeerInfoRequestType
		req = PeerInfoRequest{
			RequestID: q.id,
			PeerID:    q.key,
		}
	case ProviderQuery:
		payloadType = ProviderRequestType
		req = ProviderRequest{
			RequestID: q.id,
			Key:       q.key,
		}
	default:
		return
	}

	ctx := context.Background()
	block := blocks.NewEphemeralBlock(payloadType, req)
	block.SetHeader("requestID", q.id)
	if err := q.dht.exchange.Send(ctx, block, peersToAsk...); err != nil {
		logrus.WithError(err).Warnf("dht.next could not send block")
		return
	}
}

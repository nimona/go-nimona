package dht

import (
	"context"
	"sync"
	"time"

	"github.com/nimona/go-nimona/net"
	logrus "github.com/sirupsen/logrus"
)

const numPeersNear int = 15

type QueryType int

const (
	PeerInfoQuery QueryType = iota
	ProviderQuery
	ValueQuery
)

type query struct {
	dht            *DHT
	id             string
	key            string
	queryType      QueryType
	closestPeerID  string
	contactedPeers sync.Map
	incomingBlocks chan interface{}
	outgoingBlocks chan interface{}
}

func (q *query) Run(ctx context.Context) {
	go func() {
		// send what we know about the key
		switch q.queryType {
		case PeerInfoQuery:
			if peerInfo, err := q.dht.addressBook.GetPeerInfo(q.key); err != nil {
				q.outgoingBlocks <- peerInfo
			}
		case ProviderQuery:
			if providers, err := q.dht.store.GetProviders(q.key); err != nil {
				for _, provider := range providers {
					q.outgoingBlocks <- provider
				}
			}
		case ValueQuery:
			value, err := q.dht.store.GetValue(q.key)
			if err != nil {
				break
			}
			q.outgoingBlocks <- value
		}

		// and now, wait for something to happen
		for {
			select {
			case incomingBlock := <-q.incomingBlocks:
				switch block := incomingBlock.(type) {
				case BlockPutPeerInfoFromBlock:
					q.outgoingBlocks <- block.Peer
					// q.nextIfCloser(block.SenderPeerInfo.Metadata.Signer)
				case BlockPutProviders:
					// TODO check if id is in payload.BlockIDs
					for _, provider := range block.Providers {
						for _, blockID := range provider.Payload.(PayloadProvider).BlockIDs {
							if blockID == q.key {
								q.outgoingBlocks <- provider
								break
							}
						}
					}
					// q.nextIfCloser(block.SenderPeerInfo.Metadata.Signer)
				case BlockPutValue:
					// 	q.outgoingBlocks <- block.Value
					//  q.nextIfCloser(block.SenderPeerInfo.Metadata.Signer)
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
		if comparePeers(q.closestPeerID, newPeerID, q.key) == newPeerID {
			q.closestPeerID = newPeerID
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
		payloadType = PayloadTypeGetPeerInfo
		req = BlockGetPeerInfo{
			// SenderPeerInfo: q.dht.addressBook.GetLocalPeerInfo().Block(),
			RequestID: q.id,
			PeerID:    q.key,
		}
	case ProviderQuery:
		payloadType = PayloadTypeGetProviders
		req = BlockGetProviders{
			// SenderPeerInfo: q.dht.addressBook.GetLocalPeerInfo().Block(),
			RequestID: q.id,
			Key:       q.key,
		}
	case ValueQuery:
		payloadType = PayloadTypeGetValue
		req = BlockGetValue{
			// SenderPeerInfo: q.dht.addressBook.GetLocalPeerInfo().Block(),
			RequestID: q.id,
			Key:       q.key,
		}
	default:
		return
	}

	ctx := context.Background()
	block := net.NewEphemeralBlock(payloadType, req)
	if err := q.dht.messenger.Send(ctx, block, peersToAsk...); err != nil {
		logrus.WithError(err).Warnf("dht.next could not send block")
		return
	}
}

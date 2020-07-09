package streammanager

import (
	"sync/atomic"
	"time"

	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

type syncStatus string

const (
	syncStatusComplete  syncStatus = "COMPLETE"
	syncStatusRequested syncStatus = "REQUESTED"
)

func (m *streammanager) Sync(
	ctx context.Context,
	streamHash object.Hash,
	recipient *peer.Peer,
) (
	*Graph,
	error,
) {
	ctx = context.New(
		context.WithParent(ctx),
	)

	// only allow one sync to run at the same time for each stream
	syncAvailable := m.syncLock.TryLock(streamHash.String())
	if !syncAvailable {
		return nil, errors.New("sync for this stream is already in progress")
	}
	defer m.syncLock.Unlock(streamHash.String())

	// find the graph's knownObjects
	knownObjects, err := m.store.Filter(
		sqlobjectstore.FilterByStreamHash(streamHash),
	)
	if err != nil {
		return nil, err
	}

	// find the graph's leaves
	leafObjects := stream.GetStreamLeaves(knownObjects)
	leaves := make([]object.Hash, len(leafObjects))
	for i, lo := range leafObjects {
		leaves[i] = lo.Hash()
	}

	// setup logger
	logger := log.FromContext(ctx).With(
		log.String("method", "streammanager/streammanager.Sync"),
		log.Any("stream", streamHash),
	)

	newObjects := make(chan object.Object)

	// start listening for incoming events
	sub := m.exchange.Subscribe(
		exchange.FilterByObjectType("nimona.io/stream.**"),
	)
	defer sub.Cancel()

	missingObjectsRetrieved := make(chan struct{})
	maxMissingObjects := int64(0)
	currentMissingObjects := int64(0)

	streamRequestNonce := rand.String(12)

	go func() {
		// keep a record of all stream objects and their status
		allObjects := map[object.Hash]syncStatus{}
		// add existing objects as completed
		for _, knownObject := range knownObjects {
			allObjects[knownObject.Hash()] = syncStatusComplete
		}
		for {
			e, err := sub.Next()
			if err != nil {
				return
			}
			switch e.Payload.GetType() {
			case streamResponseType:
				p := &stream.Response{}
				err := p.FromObject(e.Payload)
				if err != nil {
					return
				}
				if !p.Stream.IsEqual(streamHash) {
					continue
				}
				// TODO start using nonces properly
				// if p.Nonce != streamRequestNonce {
				// 	continue
				// }
				logger.Debug(
					"got event list created",
					log.Any("p", p),
				)

				logger.Debug("got graph response")
				if len(p.Signatures) == 0 {
					logger.Debug("object has no signature, skipping request")
					continue
				}

				// gather the missing objects from this response
				missingObjects := []object.Hash{}
				for _, objectHash := range p.Children {
					// check sync status for object
					// and move on if completed or requested
					if _, ok := allObjects[objectHash]; ok {
						continue
					}
					// else, update the sync status
					allObjects[objectHash] = syncStatusRequested
					atomic.AddInt64(&maxMissingObjects, 1)
					atomic.AddInt64(&currentMissingObjects, 1)
					missingObjects = append(missingObjects, objectHash)
				}

				// create a request for them
				objReq := &stream.ObjectRequest{
					Nonce:   rand.String(12),
					Stream:  p.Stream,
					Objects: missingObjects,
					Owners:  m.keychain.ListPublicKeys(keychain.IdentityKey),
				}
				sig, err := object.NewSignature(
					m.keychain.GetPrimaryPeerKey(),
					objReq.ToObject(),
				)
				if err != nil {
					continue
				}
				objReq.Signatures = []object.Signature{sig}

				// and send the request to the sync response sender
				if err := m.exchange.Send(
					ctx,
					objReq.ToObject(),
					&peer.Peer{
						Owners: []crypto.PublicKey{
							e.Sender,
						},
					},
				); err != nil {
					logger.With(
						log.Any("sender", e.Sender),
						log.Error(err),
					).Debug("could not send request for stream objects")
				}

			case streamObjectResponseType:
				p := &stream.ObjectResponse{}
				err := p.FromObject(e.Payload)
				if err != nil {
					return
				}
				if !p.Stream.IsEqual(streamHash) {
					continue
				}
				// go through returned objects
				for _, obj := range p.Objects {
					if obj == nil {
						continue
					}
					obj := obj
					// check sync status for object
					// and push it to newObjects if it was not completed
					objectHash := obj.Hash()
					// TODO do we care if this was not requested?
					// TODO(geoah) who is setting this to syncStatusComplete?
					status, ok := allObjects[objectHash]
					if ok && status == syncStatusComplete {
						continue
					}
					newObjects <- *obj
					m := atomic.LoadInt64(&maxMissingObjects)
					c := atomic.AddInt64(&currentMissingObjects, -1)
					if m > 0 && c == 0 {
						missingObjectsRetrieved <- struct{}{}
					}
				}

			default:
				// if anything else, move on
				continue
			}
		}
	}()

	// create object graph request
	req := &stream.Request{
		Nonce:  streamRequestNonce,
		Stream: streamHash,
		Leaves: leaves,
		Owners: m.keychain.ListPublicKeys(keychain.IdentityKey),
	}
	sig, err := object.NewSignature(
		m.keychain.GetPrimaryPeerKey(),
		req.ToObject(),
	)
	if err != nil {
		return nil, err
	}

	req.Signatures = []object.Signature{sig}

	logger.Info("starting sync")

	logger.Debug("sending request")
	go func() {
		if err := m.exchange.Send(
			ctx,
			req.ToObject(),
			recipient,
		); err != nil {
			// TODO log error, should return if they all fail
			logger.Debug("could not send request", log.Error(err))
		}
	}()

	timeout := time.NewTimer(time.Second * 5)
loop:
	for {
		select {
		case o := <-newObjects:
			if err := m.store.Put(o); err != nil {
				logger.With(
					log.String("req.hash", streamHash.String()),
					log.Error(err),
				).Debug("could not store object")
			}
			logger.Debug(
				"got object",
				log.String("req.hash", streamHash.String()),
			)
		case <-timeout.C:
			break loop
		case <-ctx.Done():
			break loop
		case <-missingObjectsRetrieved:
			break loop
		}
	}

	// TODO currently we only support a root streams
	os, err := m.store.Filter(sqlobjectstore.FilterByStreamHash(streamHash))
	if err != nil {
		return nil, errors.Wrap(
			errors.New("could not get graph from store"),
			err,
		)
	}

	g := &Graph{
		Objects: os,
	}

	return g, nil
}

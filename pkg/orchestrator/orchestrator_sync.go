package orchestrator

import (
	"time"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

func (m *orchestrator) Sync(
	ctx context.Context,
	streamHash object.Hash,
	recipients peer.LookupOption,
) (
	*Graph,
	error,
) {
	ctx = context.New(
		context.WithParent(ctx),
	)

	// only allow one sync to run at the same time for each stream
	syncAvailable := m.syncLock.TryLock(streamHash.String())
	if syncAvailable == false {
		return nil, errors.New("sync for this stream is already in progress")
	}
	defer m.syncLock.Unlock(streamHash.String())

	// find the graph's objects
	objects, err := m.store.Filter(
		sqlobjectstore.FilterByStreamHash(streamHash),
	)
	if err != nil {
		return nil, err
	}

	// find the graph's leaves
	leafObjects := stream.GetStreamLeaves(objects)
	leaves := make([]object.Hash, len(leafObjects))
	for i, lo := range leafObjects {
		leaves[i] = hash.New(lo)
	}

	// add keys mentioned in policies
	// pks := stream.GetAllowsKeysFromPolicies(objects...)
	// for _, pk := range pks {
	// 	keys = append(keys, pk)
	// }

	// setup logger
	logger := log.FromContext(ctx).With(
		log.String("method", "orchestrator/orchestrator.Sync"),
		log.Any("stream", streamHash),
	)

	// start listening for incoming events
	newObjects := make(chan object.Object)
	streamResponse := make(chan *stream.Announcement)
	sub := m.exchange.Subscribe(exchange.FilterByObjectType("**"))
	defer sub.Cancel()
	go func() {
		for {
			e, err := sub.Next()
			if err != nil {
				return
			}
			switch e.Payload.GetType() {
			case streamAnnouncementType:
				p := &stream.Announcement{}
				err := p.FromObject(e.Payload)
				if err != nil {
					return
				}
				logger.Debug(
					"got event list created",
					log.Any("p", p),
				)
				streamResponse <- p

			default:
				newObjects <- e.Payload
			}
		}
	}()

	// net options
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithAsync(),
	}

	// keep a record of who knows about which object
	type request struct {
		hash object.Hash
		peer crypto.PublicKey
	}
	requests := make(chan *request)

	// start processing responses
	// TODO(geoah) how long should we be waiting for this part?
	// wait for ctx to timeout or for responses to come back.
	// could we wait until all requests responded or failed?
	go func() {
		respCount := 0
		for {
			select {
			case <-ctx.Done():
				// close(requests)
				return

			case res := <-streamResponse:
				logger := logger.With(
					log.Any("hashes", res.Leaves),
					log.Any("stream", res.Stream),
				)

				logger.Debug("got new events ")

				if res.Stream.String() != streamHash.String() {
					continue
				}

				logger.Debug("got graph response")
				sig := res.Signature
				for _, objectHash := range res.Leaves {
					// add a request for this hash from this peer
					if sig == nil || sig.Signer.IsEmpty() {
						logger.Debug("object has no signature, skipping request")
						continue
					}
					requests <- &request{
						hash: objectHash,
						peer: sig.Signer,
					}
				}
				respCount++
				// if respCount == len(addresses) {
				// 	// close(requests)
				// 	return
				// }
			}
		}
	}()

	go func() {
		for req := range requests {
			// check if we actually have the object
			obj, err := m.store.Get(req.hash)
			if err == nil && obj != nil {
				continue
			}

			// else we go through the peers who have it and ask them about it
			if err := m.exchange.Request(
				ctx,
				req.hash,
				peer.LookupByKey(req.peer),
				opts...,
			); err != nil {
				logger.With(
					log.Any("req.hash", req.hash),
					log.Any("req.peer", req.hash),
					log.Error(err),
				).Debug("could not send request for object")
				continue
			}
		}
	}()

	// create object graph request
	req := &stream.Request{
		Stream: streamHash,
		Leaves: leaves,
	}
	sig, err := crypto.NewSignature(
		m.localInfo.GetPeerPrivateKey(),
		req.ToObject(),
	)
	if err != nil {
		return nil, err
	}

	req.Signature = sig

	logger.Info("starting sync")

	logger.Debug("sending request")
	if err := m.exchange.Send(
		ctx,
		req.ToObject(),
		recipients,
		opts...,
	); err != nil {
		// TODO log error, should return if they all fail
		logger.Debug("could not send request")
	}

	timeout := time.NewTimer(time.Second * 5)
loop:
	for {
		select {
		case <-timeout.C:
			break loop
		case <-ctx.Done():
			break loop
		case o := <-newObjects:
			oHash := hash.New(o)
			oStreamHash := stream.GetStream(o)
			if oHash.String() != streamHash.String() {
				if oStreamHash.IsEqual(streamHash) == false {
					continue loop
				}
			}
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

func (m *orchestrator) withoutOwnAddresses(addrs []string) []string {
	clnAddrs := []string{}
	ownAddrs := m.localInfo.GetAddresses()
	skpAddrs := map[string]bool{}
	for _, o := range ownAddrs {
		skpAddrs[o] = true
	}
	for _, a := range addrs {
		if _, isOwn := skpAddrs[a]; !isOwn {
			clnAddrs = append(clnAddrs, a)
		}
	}
	return clnAddrs
}

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
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

func (m *orchestrator) Sync(
	ctx context.Context,
	streamHash object.Hash,
	addresses []string,
) (
	*Graph,
	error,
) {
	ctx = context.New(
		context.WithParent(ctx),
	)

	addresses = m.withoutOwnAddresses(addresses)

	logger := log.FromContext(ctx).With(
		log.String("method", "orchestrator/orchestrator.Sync"),
		log.Any("stream", streamHash),
		log.Strings("addresses", addresses),
	)

	// start listening for incoming events
	newObjects := make(chan object.Object)
	streamResponse := make(chan *stream.StreamResponse)
	sub := m.exchange.Subscribe(exchange.FilterByObjectType("**"))
	defer sub.Cancel()
	go func() {
		for {
			e, err := sub.Next()
			if err != nil {
				return
			}
			switch e.Payload.GetType() {
			case streamResponseType:
				p := &stream.StreamResponse{}
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
		addr string
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
					log.Any("hashes", res.Children),
					log.Any("stream", res.Stream),
				)

				logger.Debug("got new events ")

				if res.Stream.String() != streamHash.String() {
					continue
				}

				logger.Debug("got graph response")
				sig := res.Signature
				for _, objectHash := range res.Children {
					// add a request for this hash from this peer
					if sig == nil || sig.Signer.IsEmpty() {
						logger.Debug("object has no signature, skipping request")
						continue
					}
					requests <- &request{
						hash: objectHash,
						addr: "peer:" + sig.Signer.String(), // res.Identity.Fingerprint().Address(),
					}
				}
				respCount++
				if respCount == len(addresses) {
					// close(requests)
					return
				}
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
				req.addr,
				opts...,
			); err != nil {
				logger.With(
					log.Any("req.hash", req.hash),
					log.Any("req.addr", req.hash),
					log.Error(err),
				).Debug("could not send request for object")
				continue
			}
		}
	}()

	// create object graph request
	req := &stream.StreamRequest{
		Stream: streamHash,
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

	// send the request to all addresses
	for _, address := range addresses {
		logger.Debug("sending request",
			log.String("address", address),
		)
		if err := m.exchange.Send(
			ctx,
			req.ToObject(),
			address,
			opts...,
		); err != nil {
			// TODO log error, should return if they all fail
			logger.Debug("could not send request",
				log.String("address", address),
			)
		}
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
			oStreamHash := stream.Stream(o)
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

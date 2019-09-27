package orchestrator

import (
	"nimona.io/pkg/context"
	"nimona.io/internal/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/exchange"
)

func (m *orchestrator) Sync(
	ctx context.Context,
	selector []string,
	addresses []string,
) (
	*Graph,
	error,
) {
	responses := make(chan *exchange.Envelope, 10)
	addresses = m.withoutOwnAddresses(addresses)

	// create objecet graph request
	req := &ObjectGraphRequest{
		Selector: selector,
	}

	logger := log.FromContext(ctx).With(
		log.String("method", "orchestrator/orchestrator.Sync"),
		log.Strings("selector", selector),
		log.Strings("addresses", addresses),
	)

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
			exchange.WithResponse("", responses),
		); err != nil {
			// TODO log error, should return if they all fail
			logger.Debug("could not send request",
				log.String("address", address),
			)
		}
	}

	// and who knows about them
	type request struct {
		hash string
		addr string
	}
	requests := make(chan *request, 100)

	// TODO(geoah) how long should we be waiting for this part?
	// wait for ctx to timeout or for responses to come back.
	// could we wait until all requests responded or failed?
	go func() {
		respCount := 0
		for {
			select {
			case <-ctx.Done():
				close(requests)
				return

			case res := <-responses:
				logger := logger.With(
					log.String("object._hash", res.Payload.Hash().String()),
					log.String("object.type", res.Payload.GetType()),
				)

				if res.Payload.GetType() != ObjectGraphResponseType {
					continue
				}
				gres := &ObjectGraphResponse{}
				if err := gres.FromObject(res.Payload); err != nil {
					logger.Warn("could not get res from obj", log.Error(err))
					continue
				}
				logger.
					With(log.Strings("hashes", gres.ObjectHashes)).
					Debug("got graph response")
				for _, objectHash := range gres.ObjectHashes {
					// add a request for this hash from this peer
					requests <- &request{
						hash: objectHash,
						addr: "peer:" + res.Sender.Fingerprint().String(),
					}
				}
				respCount++
				if respCount == len(addresses) {
					close(requests)
					return
				}
			}
		}
	}()

	for req := range requests {
		// check if we actually have the object
		obj, err := m.store.Get(req.hash)
		if err == nil && obj != nil {
			continue
		}

		// else we go through the peers who have it and ask them about it
		out := make(chan *exchange.Envelope, 1)
		if err := m.exchange.Request(
			ctx,
			req.hash,
			req.addr,
			exchange.WithResponse("", out),
		); err != nil {
			logger.With(
				log.String("req.hash", req.hash),
				log.String("req.addr", req.hash),
				log.Error(err),
			).Debug("could not send request for object")
			close(out)
			continue
		}

		select {
		case <-ctx.Done():
		case res := <-out:
			hash := res.Payload.Hash().String()
			if hash != req.hash {
				logger.With(
					log.String("hash", hash),
					log.String("req.hash", req.hash),
					log.String("req.addr", req.hash),
					log.Error(err),
				).Debug("expected hash does not match actual")
				break
			}
			if err := m.store.Put(res.Payload); err != nil {
				logger.With(
					log.String("req.hash", req.hash),
					log.String("req.addr", req.hash),
					log.Error(err),
				).Debug("could not store objec")
			}
		}
		close(out)
	}

	// TODO currently we only support a root selector
	rootHash := selector[0]
	os, err := m.store.Graph(rootHash)
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

package main

import (
	"errors"
	"time"

	"nimona.io/internal/net"
	"nimona.io/internal/rand"
	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
)

type (
	Service struct {
		logger        log.Logger
		local         localpeer.LocalPeer
		objectmanager objectmanager.ObjectManager
		objectstore   objectstore.Store
		network       network.Network
		resolver      resolver.Resolver
		listener      net.Listener
		// configs
		config       *Config
		nimonaConfig *config.Config
	}
	Config struct {
		ReceivedFolder string `envconfig:"RECEIVED_FOLDER" default:"received_files"`
	}
)

func (srv *Service) Serve() {
	// subscribe to pin and list requests
	sub := srv.network.Subscribe(
		network.FilterByObjectType(
			objectPinRequestType,
			objectListRequestType,
		),
	)

	for e := range sub.Channel() {
		switch e.Payload.Type {
		case objectPinRequestType:
			srv.handlePinRequest(e)
		case objectListRequestType:
			srv.handleListRequest(e)
		}
	}
}

func (srv *Service) handlePinRequest(e *network.Envelope) {
	req := &PinRequest{}
	if err := req.FromObject(e.Payload); err != nil {
		return
	}

	if req.Hash.IsEmpty() {
		return
	}

	res := &PinResponse{
		Metadata: object.Metadata{
			Owner: srv.local.GetPrimaryPeerKey().PublicKey(),
		},
		RequestID: req.RequestID,
		Hash:      req.Hash,
	}

	ctx := context.New(
		context.WithTimeout(time.Second),
	)
	providers, err := srv.resolver.Lookup(
		ctx,
		resolver.LookupByContentHash(req.Hash),
	)
	if err != nil {
		res.Error = "unable to find providers for object"
	}

	for _, provider := range providers {
		ctx := context.New(
			context.WithTimeout(time.Second),
		)
		obj, err := srv.objectmanager.Request(ctx, req.Hash, provider, false)
		if err != nil {
			continue
		}
		// TODO this is a hack until we split Object TTL from Pinned flag
		srv.local.PutContentTypes(obj.Type)
		putObj, err := srv.objectmanager.Put(
			context.New(),
			obj,
		)
		if err != nil {
			continue
		}
		if err == nil {
			res.Hash = putObj.Hash()
			res.Success = true
			break
		}
	}

	if !res.Success && res.Error == "" {
		res.Error = "unable to retrieve object"
	}

	ctx = context.New(
		context.WithTimeout(time.Second),
	)
	if err := srv.network.Send(
		ctx,
		res.ToObject(),
		e.Sender,
	); err != nil {
		srv.logger.Error(
			"error sending put response",
			log.String("requestID", req.RequestID),
			log.String("sender", e.Sender.String()),
			log.Error(err),
		)
	}
}

func (srv *Service) handleListRequest(e *network.Envelope) {
	req := &ListRequest{}
	if err := req.FromObject(e.Payload); err != nil {
		return
	}

	res := &ListResponse{
		Metadata: object.Metadata{
			Owner: srv.local.GetPrimaryPeerKey().PublicKey(),
		},
		RequestID: req.RequestID,
	}

	res.Hashes, _ = srv.objectstore.GetPinned()

	ctx := context.New(
		context.WithTimeout(time.Second),
	)
	if err := srv.network.Send(
		ctx,
		res.ToObject(),
		e.Sender,
	); err != nil {
		srv.logger.Error(
			"error sending list response",
			log.String("requestID", req.RequestID),
			log.String("sender", e.Sender.String()),
			log.Error(err),
		)
	}
}

func (srv *Service) List(
	ctx context.Context,
	peerPublicKey crypto.PublicKey,
) ([]object.Hash, error) {
	peer, err := srv.resolver.LookupPeer(ctx, peerPublicKey)
	if err != nil {
		return nil, err
	}

	req := &ListRequest{
		Metadata: object.Metadata{
			Owner: srv.local.GetPrimaryPeerKey().PublicKey(),
		},
		RequestID: rand.String(8),
	}

	resChan := waitForResponse(
		ctx,
		srv.network,
		req.RequestID,
	)

	if err := srv.network.Send(
		context.New(
			context.WithTimeout(time.Second),
		),
		req.ToObject(),
		peer.PublicKey,
		network.SendWithConnectionInfo(peer),
	); err != nil {
		return nil, err
	}

	res := <-resChan
	if res == nil {
		return nil, errors.New("didn't get response in time")
	}

	listRes := &ListResponse{}
	if err := listRes.FromObject(res); err != nil {
		return nil, err
	}

	return listRes.Hashes, nil
}

func (srv *Service) Pin(
	ctx context.Context,
	peerPublicKey crypto.PublicKey,
	objHash object.Hash,
) error {
	peer, err := srv.resolver.LookupPeer(ctx, peerPublicKey)
	if err != nil {
		return err
	}

	req := &PinRequest{
		Metadata: object.Metadata{
			Owner: srv.local.GetPrimaryPeerKey().PublicKey(),
		},
		RequestID: rand.String(8),
		Hash:      objHash,
	}

	resChan := waitForResponse(
		ctx,
		srv.network,
		req.RequestID,
	)

	if err := srv.network.Send(
		context.New(
			context.WithTimeout(time.Second),
		),
		req.ToObject(),
		peer.PublicKey,
		network.SendWithConnectionInfo(peer),
	); err != nil {
		return err
	}

	res := <-resChan
	if res == nil {
		return errors.New("didn't get response in time")
	}

	pinRes := &PinResponse{}
	if err := pinRes.FromObject(res); err != nil {
		return err
	}

	if !pinRes.Success {
		if pinRes.Error != "" {
			return errors.New(pinRes.Error)
		}
		return errors.New("provider didn't pin the object")
	}

	return nil
}

func waitForResponse(
	ctx context.Context,
	net network.Network,
	rID string,
) chan *object.Object {
	sub := net.Subscribe(
		network.FilterByRequestID(rID),
	)

	res := make(chan *object.Object)

	go func() {
		defer close(res)
		defer sub.Cancel()

		done := ctx.Done()
		env := sub.Channel()

		for {
			select {
			case e := <-env:
				if e != nil && e.Payload != nil {
					res <- e.Payload
				}
				return
			case <-done:
				return
			}
		}
	}()

	return res
}

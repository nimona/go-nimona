package orchestrator

import (
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

var (
	streamRequestType  = new(stream.StreamRequest).GetType()
	streamResponseType = new(stream.StreamResponse).GetType()
)

type (
	// Orchestrator is responsible of keeping streams and their underlying
	// graphs up to date
	Orchestrator interface {
		Sync(
			ctx context.Context,
			stream object.Hash,
			addresses []string,
		) (
			*Graph,
			error,
		)
		Put(...object.Object) error
		Get(
			ctx context.Context,
			root object.Hash,
		) (
			*Graph,
			error,
		)
	}
	orchestrator struct {
		store     *sqlobjectstore.Store
		exchange  exchange.Exchange
		discovery discovery.Discoverer
		localInfo *peer.LocalPeer
	}
	Graph struct {
		Objects []object.Object
	}
)

// New constructs a new orchestrator given an object store and exchange
func New(
	store *sqlobjectstore.Store,
	exchange exchange.Exchange,
	discovery discovery.Discoverer,
	localInfo *peer.LocalPeer,
) (
	Orchestrator,
	error,
) {
	ctx := context.Background()
	return NewWithContext(
		ctx,
		store,
		exchange,
		discovery,
		localInfo,
	)
}

// NewWithContext constructs a new orchestrator given an object store and exchange
func NewWithContext(
	ctx context.Context,
	store *sqlobjectstore.Store,
	exc exchange.Exchange,
	discovery discovery.Discoverer,
	localInfo *peer.LocalPeer,
) (
	Orchestrator,
	error,
) {
	logger := log.FromContext(ctx).Named("orchestrator")
	m := &orchestrator{
		store:     store,
		exchange:  exc,
		discovery: discovery,
		localInfo: localInfo,
	}
	sub := m.exchange.Subscribe(
		exchange.FilterByObjectType("**"),
	)
	go func() {
		if err := m.process(ctx, sub); err != nil {
			logger.Error("processing failed", log.Error(err))
		}
	}()

	// Get all the content types that the local peer supports
	// find all the objects and serve only those objects
	// TODO which objects do we need to serve?
	contentTypes := m.localInfo.GetContentTypes()

	supportedObjects, err := m.store.Filter(sqlobjectstore.FilterByObjectType(contentTypes...))
	if err != nil {
		logger.Error("failed to get objects", log.Error(err))
	} else {
		// serve all the object hashes
		supportedHashes := make([]object.Hash, len(supportedObjects))

		for i, sobj := range supportedObjects {
			supportedHashes[i] = hash.New(sobj)
		}

		logger.Info(
			"adding supported object hashes as content",
			log.Any("rootObjectHashes", supportedHashes),
		)
		m.localInfo.AddContentHash(supportedHashes...)
	}

	return m, nil
}

// Process an object
func (m *orchestrator) process(ctx context.Context, sub exchange.EnvelopeSubscription) error {
	for {
		e, err := sub.Next()
		if err != nil {
			return err
		}
		ctx := context.FromContext(ctx)
		logger := log.FromContext(ctx).With(
			log.String("method", "orchestrator.Process"),
			log.String("object._hash", hash.New(e.Payload).String()),
			log.String("object.type", e.Payload.GetType()),
		)
		logger.Debug("handling object")

		o := e.Payload
		switch o.GetType() {
		case streamRequestType:
			v := &stream.StreamRequest{}
			if err := v.FromObject(o); err != nil {
				return err
			}
			if err := m.handleStreamRequest(
				ctx,
				e.Sender,
				v,
			); err != nil {
				logger.Warn("could not handle graph request object", log.Error(err))
			}
		default:
			shouldPersist := false
			for _, t := range m.localInfo.GetContentTypes() {
				if o.GetType() == t {
					shouldPersist = true
					break
				}
			}
			if !shouldPersist {
				break
			}
			if err := m.store.Put(o); err != nil {
				logger.Warn("could not persist", log.Error(err))
			}
		}
	}

	return nil
}

// IsComplete checks if a graph is missing any nodes
func IsComplete(cs []object.Object) bool {
	ms := map[string]bool{}
	cm := map[string]object.Object{}
	for _, c := range cs {
		// k: hash v: object
		cm[hash.New(c).String()] = c
	}
	for _, c := range cs {
		// get all the parents of an object
		for _, p := range stream.Parents(c) {
			h := p.String()
			// check if that hash exists in the map
			if _, ok := cm[h]; ok {
				continue
			}
			// if missing add the entry to the map
			ms[h] = true
		}
	}
	return len(ms) == 0
}

// Put stores a given object
// TODO(geoah) what happend if the graph is not complete? Error or sync?
func (m *orchestrator) Put(vs ...object.Object) error {
	hashes := make([]object.Hash, len(vs))
	for i, o := range vs {
		h := hash.New(o)
		hashes[i] = h

		// store the object
		if err := m.store.Put(o); err != nil {
			return err
		}

		// get all the objects that are part of the same graph
		os, err := m.store.Filter(sqlobjectstore.FilterByStreamHash(h))
		if err != nil {
			return errors.Wrap(
				errors.Error("could not retrieve graph"),
				err,
			)
		}

		if !IsComplete(os) {
			return errors.Wrap(
				errors.New("cannot store object"),
				ErrIncompleteGraph,
			)
		}
	}

	m.localInfo.AddContentHash(hashes...)

	return nil
}

// Get returns a complete and ordered graph given any node of the graph.
func (m *orchestrator) Get(
	ctx context.Context,
	root object.Hash,
) (
	*Graph,
	error,
) {
	os, err := m.store.Filter(sqlobjectstore.FilterByStreamHash(root))
	if err != nil {
		return nil, errors.Wrap(
			errors.Error("could not retrieve graph"),
			err,
		)
	}

	if !IsComplete(os) {
		return nil, ErrIncompleteGraph
	}

	g := &Graph{
		Objects: os,
	}

	return g, nil
}

func (m *orchestrator) handleStreamRequest(
	ctx context.Context,
	sender crypto.PublicKey,
	req *stream.StreamRequest,
) error {
	// TODO check if policy allows requested to retrieve the object
	logger := log.FromContext(ctx)

	// get the entire graph for this stream
	vs, err := m.store.Filter(sqlobjectstore.FilterByStreamHash(req.Stream))
	if err != nil {
		return err
	}

	// get only the object hashes
	hs := []object.Hash{}
	for _, o := range vs {
		hs = append(hs, hash.New(o))
	}

	res := &stream.StreamResponse{
		Stream:   req.Stream,
		Children: hs,
		Identity: m.localInfo.GetIdentityPublicKey(),
	}
	sig, err := crypto.NewSignature(
		m.localInfo.GetPeerPrivateKey(),
		req.ToObject(),
	)
	if err != nil {
		return err
	}

	res.Signature = sig

	if err := m.exchange.Send(
		ctx,
		res.ToObject(),
		"peer:"+sender.String(),
	); err != nil {
		logger.Warn(
			"orchestrator/orchestrator.handlestream.StreamRequest could not send response",
			log.Error(err),
		)
		return err
	}

	return nil
}

package orchestrator

import (
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/stream"
)

var (
	streamRequestEventListType = new(stream.RequestEventList).GetType()
	streamEventListCreatedType = new(stream.EventListCreated).GetType()
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
		store     graph.Store
		exchange  exchange.Exchange
		discovery discovery.Discoverer
		localInfo *peer.LocalPeer
		// backlog  backlog.Backlog
	}
	Graph struct {
		Objects []object.Object
	}
)

// New constructs a new orchestrator given an object store and exchange
func New(
	store graph.Store,
	exchange exchange.Exchange,
	discovery discovery.Discoverer,
	localInfo *peer.LocalPeer,
	// bc backlog.Backlog,
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
	store graph.Store,
	exc exchange.Exchange,
	discovery discovery.Discoverer,
	localInfo *peer.LocalPeer,
	// bc backlog.Backlog,
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
		// backlog:  bc,
	}
	sub := m.exchange.Subscribe(exchange.FilterByObjectType("**"))
	go func() {
		if err := m.process(ctx, sub); err != nil {
			logger.Error("processing failed", log.Error(err))
		}
	}()
	// add all local root objects to our local peer info
	// TODO should we check if these can be published?
	heads, err := m.store.Heads()
	if err != nil {
		return nil, err
	}
	rootObjectHashes := make([]object.Hash, len(heads))
	for i, rootObject := range heads {
		rootObjectHashes[i] = hash.New(rootObject)
	}
	logger.Info(
		"adding existing root object hashes as content",
		log.Any("rootObjectHashes", rootObjectHashes),
	)
	m.localInfo.AddContentHash(rootObjectHashes...)
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
		case streamRequestEventListType:
			v := &stream.RequestEventList{}
			if err := v.FromObject(o); err != nil {
				return err
			}
			if err := m.handleStreamRequestEventList(
				ctx,
				e.Sender,
				v,
			); err != nil {
				logger.Warn("could not handle graph request object", log.Error(err))
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
		cm[hash.New(c).String()] = c
	}
	for _, c := range cs {
		for _, p := range stream.Parents(c) {
			h := p.String()
			if _, ok := cm[h]; ok {
				continue
			}
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

		if err := m.store.Put(o); err != nil {
			return err
		}

		os, err := m.store.Graph(h.String())
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

	go m.localInfo.AddContentHash(hashes...)

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
	os, err := m.store.Graph(root.String())
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

func (m *orchestrator) handleStreamRequestEventList(
	ctx context.Context,
	sender crypto.PublicKey,
	req *stream.RequestEventList,
) error {
	// TODO check if policy allows requested to retrieve the object
	logger := log.FromContext(ctx)

	vs, err := m.store.Graph(req.Stream.String())
	if err != nil {
		return err
	}

	hs := []object.Hash{}
	for _, o := range vs {
		hs = append(hs, hash.New(o))
	}

	res := &stream.EventListCreated{
		Stream:   req.Stream,
		Events:   hs,
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
			"orchestrator/orchestrator.handlestream.RequestEventList could not send response",
			log.Error(err),
		)
		return err
	}

	return nil
}

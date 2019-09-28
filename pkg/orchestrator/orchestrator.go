package orchestrator

import (
	"fmt"

	"nimona.io/internal/store/graph"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/stream"
)

var (
	streamRequestEventListType = new(stream.RequestEventList).GetType()
	streamEventListCreatedType = new(stream.EventListCreated).GetType()
)

//go:generate $GOBIN/genny -in=$GENERATORS/pubsub/pubsub.go -out=pubsub_string_generated.go -pkg orchestrator gen "ObservableType=string"

type (
	// Orchestrator is responsible of keeping streams and their underlying
	// graphs up to date
	Orchestrator interface {
		Subscriber
		Sync(
			ctx context.Context,
			selector []string,
			addresses []string,
		) (
			*Graph,
			error,
		)
		Put(...object.Object) error
		Get(
			ctx context.Context,
			rootHash string,
		) (
			*Graph,
			error,
		)
	}
	orchestrator struct {
		PubSub
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
	exchange exchange.Exchange,
	discovery discovery.Discoverer,
	localInfo *peer.LocalPeer,
	// bc backlog.Backlog,
) (
	Orchestrator,
	error,
) {
	m := &orchestrator{
		PubSub:    NewPubSub(),
		store:     store,
		exchange:  exchange,
		discovery: discovery,
		localInfo: localInfo,
		// backlog:  bc,
	}
	if _, err := m.exchange.Handle("**", m.Process); err != nil {
		return nil, err
	}
	// add all local root objects to our local peer info
	// TODO should we check if these can be published?
	heads, err := m.store.Heads()
	if err != nil {
		return nil, err
	}
	rootObjectHashes := make([]string, len(heads))
	for i, rootObject := range heads {
		rootObjectHashes[i] = rootObject.Hash().String()
	}
	m.localInfo.AddContentHash(rootObjectHashes...)
	return m, nil
}

// Process an object
func (m *orchestrator) Process(e *exchange.Envelope) error {
	ctx := context.Background()
	logger := log.FromContext(ctx).With(
		log.String("object._hash", e.Payload.Hash().String()),
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
		reqID := o.Get(exchange.ObjectRequestID).(string)
		if err := m.handleStreamRequestEventList(
			ctx,
			reqID,
			e.Sender,
			v,
		); err != nil {
			logger.Warn("could not handle graph request object", log.Error(err))
		}
	}

	return nil
}

// IsComplete checks if a graph is missing any nodes
func IsComplete(cs []object.Object) bool {
	ms := map[string]bool{}
	cm := map[string]object.Object{}
	for _, c := range cs {
		cm[c.Hash().String()] = c
	}
	for _, c := range cs {
		for _, p := range c.GetParents() {
			if _, ok := cm[p]; ok {
				continue
			}
			ms[p] = true
		}
	}
	return len(ms) == 0
}

// Put stores a given object
// TODO(geoah) what happend if the graph is not complete? Error or sync?
func (m *orchestrator) Put(vs ...object.Object) error {
	hashes := make([]string, len(vs))
	for i, o := range vs {
		hashes[i] = o.Hash().String()

		if err := m.store.Put(o); err != nil {
			return err
		}

		os, err := m.store.Graph(o.Hash().String())
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

		m.Publish(o.Hash().String())
	}

	go m.localInfo.AddContentHash(hashes...)

	return nil
}

// Get returns a complete and ordered graph given any node of the graph.
func (m *orchestrator) Get(
	ctx context.Context,
	rootHash string,
) (
	*Graph,
	error,
) {
	os, err := m.store.Graph(rootHash)
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
	reqID string,
	sender *crypto.PublicKey,
	req *stream.RequestEventList,
) error {
	// TODO check if policy allows requested to retrieve the object
	logger := log.FromContext(ctx)

	vs, err := m.store.Graph(req.StreamHashes[0])
	if err != nil {
		return err
	}

	hs := []string{}
	for _, o := range vs {
		hs = append(hs, o.Hash().String())
	}

	res := &stream.EventListCreated{
		EventHashes: hs,
	}

	if err := m.exchange.Send(
		ctx,
		res.ToObject(),
		"peer:"+sender.Fingerprint().String(),
		exchange.AsResponse(reqID),
	); err != nil {
		logger.Warn(
			"orchestrator/orchestrator.handlestream.RequestEventList could not send response",
			log.Error(err),
		)
		return err
	}

	return nil
}

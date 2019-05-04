package dag

import (
	"time"

	"go.uber.org/zap"
	"nimona.io/internal/context"
	"nimona.io/internal/errors"
	"nimona.io/internal/log"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/exchange"
)

//go:generate go run github.com/cheekybits/genny -in=../../../internal/generator/pubsub/pubsub.go -out=pubsub_string_generated.go -pkg dag gen "ObservableType=string"
//TODO go:generate go run github.com/cheekybits/genny -in=../../../internal/generator/queue/queue.go -out=queue_object_generated.go -extra-imports "nimona.io/pkg/object" -pkg dag gen "ObservableType=*object.Object"

type (
	// Manager is responsible of keeping track of all the objects, graphs,
	// and mutations, and exposing the to the clients
	Manager interface {
		Subscriber
		Sync(
			ctx context.Context,
			selector []string,
			addresses []string,
		) (
			[]*object.Object,
			error,
		)
		Put(...*object.Object) error
		Get(
			ctx context.Context,
			rootHash string,
		) (
			[]*object.Object,
			error,
		)
	}
	manager struct {
		PubSub
		store     graph.Store
		exchange  exchange.Exchange
		discovery discovery.Discoverer
		localInfo *net.LocalInfo
		// backlog  backlog.Backlog
	}
	// options for Manager.Get()
	getOptions struct {
		request bool
		timeout time.Duration
	}
	getOption func(*getOptions)
)

// New constructs a new manager given an object store and exchange
func New(
	store graph.Store,
	exchange exchange.Exchange,
	discovery discovery.Discoverer,
	localInfo *net.LocalInfo,
	// bc backlog.Backlog,
) (
	Manager,
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

// NewWithContext constructs a new manager given an object store and exchange
func NewWithContext(
	ctx context.Context,
	store graph.Store,
	exchange exchange.Exchange,
	discovery discovery.Discoverer,
	localInfo *net.LocalInfo,
	// bc backlog.Backlog,
) (
	Manager,
	error,
) {
	m := &manager{
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
		rootObjectHashes[i] = rootObject.HashBase58()
	}
	m.localInfo.AddContentHash(rootObjectHashes...)
	return m, nil
}

// Process an object
func (m *manager) Process(e *exchange.Envelope) error {
	ctx := context.Background()
	logger := log.Logger(ctx).With(
		zap.String("object._hash", e.Payload.HashBase58()),
		zap.String("object.type", e.Payload.GetType()),
	)
	logger.Debug("handling object")

	o := e.Payload
	switch o.GetType() {
	case ObjectGraphRequestType:
		v := &ObjectGraphRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		reqID := o.GetRaw(exchange.ObjectRequestID).(string)
		if err := m.handleObjectGraphRequest(
			ctx,
			reqID,
			e.Sender,
			v,
		); err != nil {
			logger.Warn("could not handle graph request object", zap.Error(err))
		}
	}

	return nil
}

// IsComplete checks if a graph is missing any nodes
func IsComplete(cs []*object.Object) bool {
	ms := map[string]bool{}
	cm := map[string]*object.Object{}
	for _, c := range cs {
		cm[c.HashBase58()] = c
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
func (m *manager) Put(vs ...*object.Object) error {
	hashes := make([]string, len(vs))
	for i, o := range vs {
		hashes[i] = o.HashBase58()

		if err := m.store.Put(o); err != nil {
			return err
		}

		os, err := m.store.Graph(o.HashBase58())
		if err != nil {
			return errors.Wrap(
				errors.Error("could not retrieve graph"),
				err,
			)
		}

		if !IsComplete(os) {
			return ErrIncompleteGraph
		}

		m.Publish(o.HashBase58())
	}

	m.localInfo.AddContentHash(hashes...)

	return nil
}

// Get returns a complete and ordered graph given any node of the graph.
func (m *manager) Get(
	ctx context.Context,
	rootHash string,
) (
	[]*object.Object,
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

	return os, nil
}

func (m *manager) handleObjectGraphRequest(
	ctx context.Context,
	reqID string,
	sender *crypto.PublicKey,
	req *ObjectGraphRequest,
) error {
	// TODO check if policy allows requested to retrieve the object
	logger := log.Logger(ctx)

	vs, err := m.store.Graph(req.Selector[0])
	if err != nil {
		return err
	}

	hs := []string{}
	for _, o := range vs {
		hs = append(hs, o.HashBase58())
	}

	res := &ObjectGraphResponse{
		ObjectHashes: hs,
	}

	if err := m.exchange.Send(
		ctx,
		res.ToObject(),
		"peer:"+sender.Fingerprint(),
		exchange.AsResponse(reqID),
	); err != nil {
		logger.Warn(
			"dag/manager.handleObjectGraphRequest could not send response",
			zap.Error(err),
		)
		return err
	}

	return nil
}

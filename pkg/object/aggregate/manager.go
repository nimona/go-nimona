package aggregate

import (
	"time"

	"nimona.io/internal/context"
	"nimona.io/internal/errors"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/dag"
	"nimona.io/pkg/object/exchange"
	"nimona.io/pkg/object/mutation"
)

//go:generate go run github.com/cheekybits/genny -in=../../../internal/generator/pubsub/pubsub.go -out=pubsub_generate.go -pkg aggregate gen "ObservableType=*AggregateObject"

type (
	// Manager for object aggregates
	Manager interface {
		Subscriber
		Append(string, ...*mutation.Operation) error
		Get(context.Context, string) (*AggregateObject, error)
	}
	// manager implementation
	manager struct {
		PubSub
		exchange     exchange.Exchange
		store        graph.Store
		dag          dag.Manager
		aggregator   *Aggregator
		graphUpdated chan string
	}
	// options for Manager.Get()
	getOptions struct {
		sync    bool
		timeout time.Duration
	}
	getOption func(*getOptions)
)

// New constructs a new manager given an object store and exchanges
func New(
	store graph.Store,
	exchange exchange.Exchange,
	dag dag.Manager,
) (
	Manager,
	error,
) {
	m := &manager{
		PubSub:       NewPubSub(),
		exchange:     exchange,
		store:        store,
		dag:          dag,
		aggregator:   NewAggregator(),
		graphUpdated: make(chan string, 100),
	}
	dag.Subscribe(m.graphUpdated)
	go m.process()
	return m, nil
}

func WithTimeout(t time.Duration) getOption {
	return func(o *getOptions) {
		o.timeout = t
	}
}

// Get returns an aggregated object given a base object hash.
func (m *manager) Get(ctx context.Context, hash string) (*AggregateObject, error) {
	os, err := m.dag.Get(ctx, hash)
	if err != nil {
		return nil, errors.Wrap(
			errors.Error("could not get complete graph"),
			err,
		)
	}

	var ro *object.Object
	ms := []*mutation.Mutation{}
	for i := range os {
		if os[i].GetType() != mutation.MutationType {
			if ro != nil {
				return nil, errors.Error("more than one basic objects")
			}
			ro = os[i]
			continue
		}
		m := &mutation.Mutation{}
		if mErr := m.FromObject(os[i]); mErr != nil {
			return nil, errors.Wrap(
				errors.New("could not retrieve from object"), mErr)
		}
		ms = append(ms, m)
	}

	ao, err := m.aggregator.Aggregate(ro, ms)
	if err != nil {
		// TODO wrap log and possibly handle error
		return nil, err
	}

	return ao, nil
}

func (m *manager) process() {
	for {
		select {
		case h := <-m.graphUpdated:
			// TODO This is way too expensive if no one actually cares about
			// this aggregate.
			ctx := context.Background()
			a, err := m.Get(ctx, h)
			if err != nil {
				// TODO log error
				continue
			}
			m.Publish(a)
		}
	}
}

// Append looks for the tail of a graph given its root hash, and created a
// mutation using them as parents.
// TODO(geoah) The mutation created using Append in the tests has a different
// hash than expected. Not sure why.
func (m *manager) Append(
	rootHash string,
	ops ...*mutation.Operation,
) error {
	// find root object
	o, err := m.store.Get(rootHash)
	if err != nil {
		return errors.Wrap(
			errors.Error("could not get graph root"),
			err,
		)
	}

	// find graph tail
	ts, err := m.store.Tails(o.HashBase58())
	if err != nil {
		return errors.Wrap(
			errors.Error("could not get graph tails"),
			err,
		)
	}

	if len(ts) == 0 {
		return errors.Error("did not find graph tails")
	}

	// create mutation
	ps := []string{}
	for _, t := range ts {
		ps = append(ps, t.HashBase58())
	}
	u := mutation.New(ops, ps)

	// put mutation
	if err := m.dag.Put(u.ToObject()); err != nil {
		return errors.Wrap(
			errors.Error("could not store mutation"),
			err,
		)
	}

	// TODO(geoah) return aggregate?

	return nil
}

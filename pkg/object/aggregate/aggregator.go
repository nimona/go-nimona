package aggregate

import (
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/mutation"
)

type (
	// Aggregator deals with objects and their mutations
	// TODO consider introducing some sort of cache
	Aggregator struct{}
	// AggregateObject is an object with its mutations applied
	AggregateObject struct {
		Aggregate *object.Object
		Root      *object.Object
		Mutations []*mutation.Mutation
	}
)

// GetType returns the object's type
func (o *AggregateObject) GetType() string {
	return o.Root.GetType()
}

// NewAggregator constructs a new aggregator
func NewAggregator() *Aggregator {
	return &Aggregator{}
}

// Mutate aggregate object
func (o *AggregateObject) Mutate(m *mutation.Mutation) error {
	// TODO Check all deps (parents) of the mutation have been allready applied
	// TODO add mutation to aggregate's applied mutations
	return m.Mutate(o.Aggregate)
}

// Aggregate given an ordered list of objects and mutations,
// returns a list of aggregate objects with their mutations applied.
func (a *Aggregator) Aggregate(
	o *object.Object,
	ms []*mutation.Mutation,
) (
	*AggregateObject,
	error,
) {
	ao := &AggregateObject{
		Root:      o.Copy(),
		Aggregate: o.Copy(),
		Mutations: []*mutation.Mutation{},
	}
	for _, m := range ms {
		if err := m.Mutate(ao.Aggregate); err != nil {
			// TODO Log error
			continue
		}
	}
	return ao, nil
}

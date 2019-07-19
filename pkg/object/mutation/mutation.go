package mutation

import (
	"nimona.io/pkg/object"
)

//go:generate $GOBIN/objectify -schema /object.mutation -type Mutation -in mutation.go -out mutation_generated.go

// Mutation provides operations to be applied given that the parent mutations
// have already been applied
type Mutation struct {
	Operations []*Operation `json:"ops:ao"`
	Parents    []string     `json:"@parents:as"`
}

// New construct a mutation from an array of operations and parrent chains
func New(ops []*Operation, parents []string) *Mutation {
	return &Mutation{
		Operations: ops,
		Parents:    parents,
	}
}

// Mutate applies the mutation's operations on the given object
func (c Mutation) Mutate(o object.Object) error {
	for _, operation := range c.Operations {
		if err := operation.Apply(o); err != nil {
			return err
		}
	}
	return nil
}

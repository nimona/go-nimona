package graph

import (
	"nimona.io/pkg/object"
)

// Store interface
type Store interface {
	Put(*object.Object) error
	Get(string) (*object.Object, error)
	Graph(string) ([]*object.Object, error)
	Children(string) ([]*object.Object, error)
	Head(string) (*object.Object, error)
	Heads() ([]*object.Object, error)
	Tails(string) ([]*object.Object, error)
}

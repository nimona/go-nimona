package rel

import (
	"nimona.io/pkg/object"
)

type ObjectStore interface {
	Put(object.Object, ...Option) error
	Get(object.Hash) (object.Object, error)
	UpdateTTL(hash object.Hash, minutes int) error
	Delete(object.Hash) error
}

type StreamStore interface {
	PutRelation(parent object.Hash, child object.Hash, options ...Option) error
	GetRelations(parent object.Hash) ([]object.Hash, error)
	Subscribe(parent object.Hash) (chan object.Hash, error)
}

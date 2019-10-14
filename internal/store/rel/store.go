package rel

import "nimona.io/pkg/object"

type Store interface {
	Store(obj object.Object, ttl int) error
	GetByHash(hash string) (object.Object, error)
	GetByStreamHash(streamHash string) (object.Object, error)
	UpdateTTL(hash string, minutes int) error
}

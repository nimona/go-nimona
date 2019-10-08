package rel

import "nimona.io/pkg/object"

type Store interface {
	Get(string) (object.Object, error)
	Create(object.Object) error
}

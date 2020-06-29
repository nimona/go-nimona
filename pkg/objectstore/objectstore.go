package objectstore

import "nimona.io/pkg/object"

//go:generate $GOBIN/mockgen -destination=./objectstoremock/objectstoremock_generated.go -package=objectstoremock -source=objectstore.go

type (
	Getter interface {
		Get(
			hash object.Hash,
		) (object.Object, error)
	}
	Store interface {
		Getter
	}
)

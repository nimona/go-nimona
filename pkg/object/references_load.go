package object

import (
	"github.com/hashicorp/go-multierror"

	"nimona.io/pkg/context"
)

// LoadReferences will look for references in the given object, request the
// referred objects using the getter, and will replace the references
// with the actual object before returning the complete
func LoadReferences(
	ctx context.Context,
	objectCID CID,
	getter GetterFunc,
) (*Object, error) {
	obj, err := getter(
		ctx,
		objectCID,
	)
	if err != nil {
		return nil, err
	}
	var getError error
	traverseObject(obj, func(k string, v Value) (string, Value, bool) {
		switch vv := v.(type) {
		case CID:
			o, err := getter(ctx, vv)
			if err != nil {
				getError = multierror.Append(getError, err)
				return "", nil, false
			}
			return k, o, true
		case CIDArray:
			// TODO implement and test
			panic("LoadReferences doesn't implement loading from CIDArray")
		}
		return "", nil, false
	})
	return obj, nil
}

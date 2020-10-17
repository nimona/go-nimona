package object

import (
	"strings"

	"github.com/hashicorp/go-multierror"

	"nimona.io/pkg/context"
)

// LoadReferences will look for references in the given object, request the
// referred objects using the getter, and will replace the references
// with the actual object before returning the complete
func LoadReferences(
	ctx context.Context,
	objectHash Hash,
	getter GetterFunc,
) (*Object, error) {
	obj, err := getter(
		ctx,
		objectHash,
	)
	if err != nil {
		return nil, err
	}
	var getError error
	traverseObject(obj, func(k string, v interface{}) (string, interface{}, bool) {
		h, ok := v.(Hash)
		if !ok {
			return "", nil, false
		}
		switch {
		case strings.HasSuffix(k, ":r"):
			o, err := getter(ctx, h)
			if err != nil {
				getError = multierror.Append(getError, err)
				return "", nil, false
			}
			return strings.Replace(k, ":r", ":m", 1), o, true
		case strings.HasSuffix(k, ":am"):
			panic("LoadReferences doesn't implement loading from slices")
		}
		return "", nil, false
	})
	return obj, nil
}

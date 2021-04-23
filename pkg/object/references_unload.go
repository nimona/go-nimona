package object

import (
	"nimona.io/pkg/context"
)

func UnloadReferences(
	ctx context.Context,
	obj *Object,
) (
	result *Object,
	unloaded []*Object,
	err error,
) {
	traverseObject(obj, func(k string, v Value) (string, Value, bool) {
		switch vv := v.(type) {
		case *Object:
			unloaded = append(unloaded, vv)
			return k, vv.CID(), true
		case ObjectArray:
			hs := CIDArray{}
			for _, o := range vv {
				unloaded = append(unloaded, o)
				hs = append(hs, o.CID())
			}
			return k, hs, true
		}
		return "", nil, false
	})
	return obj, unloaded, nil
}

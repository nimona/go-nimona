package object

import (
	"strings"

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
	traverseObject(obj, func(k string, v interface{}) (string, interface{}, bool) {
		switch {
		case strings.HasSuffix(k, ":m"):
			o, ok := v.(*Object)
			if !ok {
				return "", nil, false
			}
			unloaded = append(unloaded, o)
			return strings.Replace(k, ":m", ":r", 1), o.Hash(), true
		case strings.HasSuffix(k, ":o"):
			o, ok := v.(*Object)
			if !ok {
				return "", nil, false
			}
			unloaded = append(unloaded, o)
			return strings.Replace(k, ":o", ":r", 1), o.Hash(), true
		case strings.HasSuffix(k, ":ao"):
			switch vs := v.(type) {
			case []*Object:
				hs := []Hash{}
				for _, o := range vs {
					unloaded = append(unloaded, o)
					hs = append(hs, o.Hash())
				}
				return strings.Replace(k, ":ao", ":ar", 1), hs, true
			case []interface{}:
				hs := []Hash{}
				for _, vsv := range vs {
					o, ok := vsv.(*Object)
					if !ok {
						continue
					}
					unloaded = append(unloaded, o)
					hs = append(hs, o.Hash())
				}
				if len(hs) > 0 {
					return strings.Replace(k, ":ao", ":ar", 1), hs, true
				}
			}
		}
		return "", nil, false
	})
	return obj, unloaded, nil
}

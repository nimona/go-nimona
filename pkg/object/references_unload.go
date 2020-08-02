package object

import (
	"strings"

	"nimona.io/pkg/context"
)

func UnloadReferences(
	ctx context.Context,
	obj Object,
) (
	result *Object,
	unloaded []Object,
	err error,
) {
	objs := map[string]Hash{}
	data := obj.Raw().Value("data:m")
	Traverse(data, func(k string, v Value) bool {
		if !v.IsMap() {
			return true
		}
		t := v.(Map).Value("type:s")
		if t == nil {
			return true
		}
		objs[k] = v.Hash()
		unloaded = append(unloaded, Object(v.(Map)))
		return true
	})
	for k, ref := range objs {
		obj = obj.Set(k, nil)
		nk := strings.Replace(k, ":m", ":r", 1)
		obj = obj.Set(nk, Ref(ref))
	}
	return &obj, unloaded, nil
}

package object

import (
	"strings"

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
	refs := map[string]Hash{}
	data := obj.Raw().Value("data:m")
	Traverse(data, func(k string, v Value) bool {
		if !v.IsRef() {
			return true
		}
		refs[k] = Hash(v.(Ref))
		return true
	})
	refObjs := map[string]*Object{}
	for k, ref := range refs {
		refObj, err := getter(ctx, ref)
		if err != nil {
			return nil, err
		}
		refObjs[k] = refObj
	}
	fullObj := *obj
	for k, refObj := range refObjs {
		fullObj = fullObj.Set(k, nil)
		nk := strings.Replace(k, ":r", ":m", 1)
		fullObj = fullObj.Set(nk, refObj.ToMap())
	}
	return &fullObj, nil
}

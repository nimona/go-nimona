package objectloader

import (
	"strings"

	"nimona.io/pkg/exchange"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/resolver"
)

type (
	Loader interface {
		Load(object.Object, ...option) (object.Object, error)
		Unload(object.Object, ...option) (object.Object, error)
	}
	loader struct {
		store    objectstore.Store
		resolver resolver.Resolver
		exchange exchange.Exchange
	}
	option  func(*options)
	options struct {
		loadTypes []string
		skipTypes []string
	}
)

func WithType(t string) func(*options) {
	return func(o *options) {
		o.loadTypes = append(o.loadTypes, t)
	}
}

func WithoutType(t string) func(*options) {
	return func(o *options) {
		o.skipTypes = append(o.skipTypes, t)
	}
}

func (l *loader) Load(
	obj object.Object,
	opts ...option,
) (*object.Object, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	refs := map[string]object.Hash{}
	data := obj.Raw().Value("data:m")
	object.Traverse(data, func(k string, v object.Value) bool {
		if !v.IsRef() {
			return true
		}
		refs[k] = object.Hash(v.(object.Ref))
		return true
	})
	refObjs := map[string]object.Object{}
	for k, ref := range refs {
		refObj, err := l.store.Get(ref)
		if err != nil {
			return nil, err
		}
		refObjs[k] = refObj
	}
	for k, refObj := range refObjs {
		obj = obj.Set(k, nil)
		nk := strings.Replace(k, ":r", ":m", 1)
		obj = obj.Set(nk, refObj.ToMap())
	}
	return &obj, nil
}

func (l *loader) Unload(
	obj object.Object,
	opts ...option,
) (*object.Object, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	objs := map[string]object.Hash{}
	data := obj.Raw().Value("data:m")
	object.Traverse(data, func(k string, v object.Value) bool {
		if !v.IsMap() {
			return true
		}
		t := v.(object.Map).Value("type:s")
		if t == nil {
			return true
		}
		objs[k] = v.Hash()
		return true
	})
	for k, ref := range objs {
		obj = obj.Set(k, nil)
		nk := strings.Replace(k, ":m", ":r", 1)
		obj = obj.Set(nk, object.Ref(ref))
	}
	return &obj, nil
}

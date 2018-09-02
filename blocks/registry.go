package blocks

import (
	"reflect"
	"sync"
)

func ParseRegistryOptions(opts ...RegistryOption) *RegistryOptions {
	options := &RegistryOptions{}
	for _, o := range opts {
		o(options)
	}
	return options
}

type RegistryOptions struct {
	Persist bool
}

type RegistryOption func(*RegistryOptions)

func Persist() RegistryOption {
	return func(opts *RegistryOptions) {
		opts.Persist = true
	}
}

var registry = &Registry{
	types:   sync.Map{},
	persist: sync.Map{},
}

// RegisterContentType registers types and content on a default registry
func RegisterContentType(contentType string, content interface{}, opts ...RegistryOption) {
	registry.Register(contentType, content, opts...)
}

func ShouldPersist(contentType string) bool {
	_, ok := registry.persist.Load(contentType)
	return ok
}

func GetFromType(t reflect.Type) string {
	return registry.GetFromType(t)
}

func GetType(contentType string) reflect.Type {
	return registry.GetType(contentType)
}

// Registry holds content types and their structures
type Registry struct {
	types   sync.Map
	persist sync.Map
}

// Register a content type and its structure
func (r *Registry) Register(contentType string, content interface{}, opts ...RegistryOption) {
	options := ParseRegistryOptions(opts...)
	r.types.Store(contentType, reflect.TypeOf(content))
	if options.Persist {
		registry.persist.Store(contentType, true)
	}
}

func (r *Registry) GetType(contentType string) reflect.Type {
	t, ok := r.types.Load(contentType)
	if !ok {
		return nil
	}
	return t.(reflect.Type)
}

func (r *Registry) GetFromType(t reflect.Type) string {
	var rt string
	r.types.Range(func(k, v interface{}) bool {
		if v.(reflect.Type).String() == t.String() {
			rt = k.(string)
		}
		return true
	})
	return rt
}

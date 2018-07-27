package net

import (
	"reflect"
	"sync"
)

var registry = &Registry{
	types: sync.Map{},
}

// RegisterContentType registers types and content on a default registry
func RegisterContentType(contentType string, content interface{}) {
	registry.Register(contentType, content)
}

// GetContentType returns a content type's structure from a default registry
// TODO Bad function name
func GetContentType(contentType string) interface{} {
	return registry.Get(contentType)
}

// Registry holds content types and their structures
type Registry struct {
	types sync.Map
}

// Register a content type and its structure
func (r *Registry) Register(contentType string, content interface{}) {
	r.types.Store(contentType, reflect.TypeOf(content))
}

// Get returns a content type's structure
func (r *Registry) Get(contentType string) interface{} {
	t, ok := r.types.Load(contentType)
	if !ok {
		return map[string]interface{}{}
	}

	v := reflect.New(t.(reflect.Type)).Elem()
	return v.Interface()
}

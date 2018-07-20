package net

import (
	"reflect"
	"sync"
)

var registry = &Registry{
	types: map[string]reflect.Type{},
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
	lock  sync.RWMutex
	types map[string]reflect.Type
}

// Register a content type and its structure
func (r *Registry) Register(contentType string, content interface{}) {
	r.lock.Lock()
	r.types[contentType] = reflect.TypeOf(content)
	r.lock.Unlock()
}

// GetContentType returns a content type's structure
func (r *Registry) Get(contentType string) interface{} {
	r.lock.RLock()
	t, ok := r.types[contentType]
	if !ok {
		return map[string]interface{}{}
	}
	v := reflect.New(t).Elem()
	r.lock.RUnlock()
	return v.Interface()
}

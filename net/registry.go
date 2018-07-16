package net

import (
	"reflect"
	"sync"
)

var registry = &Registry{
	types: map[string]reflect.Type{},
}

func RegisterContentType(contentType string, content interface{}) {
	registry.Register(contentType, content)
}

func GetContentType(contentType string) interface{} {
	return registry.Get(contentType)
}

type Registry struct {
	lock  sync.RWMutex
	types map[string]reflect.Type
}

func (r *Registry) Register(contentType string, content interface{}) {
	r.lock.Lock()
	r.types[contentType] = reflect.TypeOf(content)
	r.lock.Unlock()
}

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

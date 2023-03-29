package kv

import (
	"fmt"
	"reflect"
	"strings"
)

type Store[K, V any] interface {
	Set(K, *V) error
	Get(K) (*V, error)
	GetPrefix(K) ([]*V, error)
}

func keyToString(key interface{}) string {
	if str, ok := key.(interface{ String() string }); ok {
		return str.String()
	}
	switch reflect.TypeOf(key).Kind() {
	case reflect.Struct:
		keys := []string{}
		val := reflect.ValueOf(key)
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if !field.CanInterface() {
				continue
			}
			part := fmt.Sprintf("%v", field.Interface())
			if part != "" {
				keys = append(keys, part)
			}
		}
		return strings.Join(keys, "/") + "/"
	case reflect.Slice:
		keys := []string{}
		val := reflect.ValueOf(key)
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			part := fmt.Sprintf("%v", elem.Interface())
			if part != "" {
				keys = append(keys, part)
			}
		}
		return strings.Join(keys, "/") + "/"
	default:
		return fmt.Sprintf("%v", key) + "/"
	}
}

package object

import (
	"reflect"
	"strconv"
	"strings"
)

func traverseObject(
	v Value,
	f func(string, Value) (updatedKey string, updatedValue Value, update bool),
) {
	switch vv := v.(type) {
	case *Object:
		traverseObject(vv.Data, f)
	case Map:
		for ik, iv := range vv {
			uk, uv, u := f(ik, iv)
			if u {
				vv[uk] = uv
				if ik != uk {
					delete(vv, ik)
				}
				traverseObject(uv, f)
			} else {
				traverseObject(iv, f)
			}
		}
	}
}

// nolint: unused
func traverseValues(v reflect.Value, f func(reflect.Value)) {
	f(v)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for j := 0; j < v.Len(); j++ {
			traverseValues(v.Index(j), f)
		}
	case reflect.Map:
		for _, vk := range v.MapKeys() {
			traverseValues(v.MapIndex(vk), f)
		}
	case reflect.Struct:
		for j := 0; j < v.NumField(); j++ {
			vf := v.Field(j)
			traverseValues(vf, f)
		}
	}
}

func traverse(k string, i interface{}, f func(string, interface{}) bool) bool {
	cont := f(k, i)
	if !cont {
		return false
	}
	if _, ok := i.([]byte); ok {
		return true
	}
	v := reflect.ValueOf(i)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for j := 0; j < v.Len(); j++ {
			kk := strings.Trim(k+"/"+strconv.Itoa(j), "/")
			cont = traverse(kk, v.Index(j).Interface(), f)
			if !cont {
				return false
			}
		}
	case reflect.Map:
		for _, vk := range v.MapKeys() {
			kk := strings.Trim(k+"/"+vk.Interface().(string), "/")
			cont = traverse(kk, v.MapIndex(vk).Interface(), f)
			if !cont {
				return false
			}
		}
	case reflect.Struct:
		for j := 0; j < v.NumField(); j++ {
			vf := v.Field(j)
			// !CanInterface() is a workaround for:
			// `cannot return value obtained from unexported field or method`
			if !vf.CanInterface() {
				return true
			}
			kk := strings.Trim(k+"/"+vf.Type().Name(), "/")
			cont = traverse(kk, vf.Interface(), f)
			if !cont {
				return false
			}
		}
	}
	return true
}

func Traverse(v interface{}, f func(string, interface{}) bool) {
	traverse("", v, f)
}

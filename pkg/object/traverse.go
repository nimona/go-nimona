package object

import (
	"reflect"
	"strconv"
)

func traverseObject(
	v interface{},
	f func(string, interface{}) (string, interface{}, bool),
) {
	switch vv := v.(type) {
	case *Object:
		traverseObject(vv.Data, f)
	case map[string]interface{}:
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
			cont = traverse(k+strconv.Itoa(j), v.Index(j).Interface(), f)
			if !cont {
				return false
			}
		}
	case reflect.Map:
		for _, vk := range v.MapKeys() {
			cont = traverse(k+vk.Interface().(string), v.MapIndex(vk).Interface(), f)
			if !cont {
				return false
			}
		}
	case reflect.Struct:
		for j := 0; j < v.NumField(); j++ {
			vf := v.Field(j)
			cont = traverse(k+vf.Type().Name(), vf.Interface(), f)
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

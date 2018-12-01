package encoding

// import (
// 	"reflect"

// 	"github.com/mitchellh/mapstructure"
// )

// // Encode is a wrapper for mapstructure's Decode with our decodeHook that allows
// // encoding structs to maps
// // TODO move addCtx to an option
// func Encode(from interface{}, addCtx bool) (map[string]interface{}, error) {
// 	m := map[string]interface{}{}
// 	dc := &mapstructure.DecoderConfig{
// 		Metadata:         &mapstructure.Metadata{},
// 		DecodeHook:       getEncodeHook(addCtx),
// 		Result:           &m,
// 		TagName:          "json",
// 		WeaklyTypedInput: true,
// 	}
// 	dec, err := mapstructure.NewDecoder(dc)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := dec.Decode(from); err != nil {
// 		return nil, err
// 	}

// 	// spew.Dump(m)

// 	// tm, err := TypeMap(m)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	return m, nil
// }

// func getEncodeHook(addCtx bool) mapstructure.DecodeHookFuncType {
// 	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
// 		// fmt.Println("------------")
// 		// fmt.Println("from", from, from.Kind())
// 		// fmt.Println("to", to, to.Kind())
// 		// fmt.Println("data", data)

// 		if !addCtx {
// 			// TODO(geoah) explain this hack
// 			addCtx = true
// 			return data, nil
// 		}

// 		// decoding registered struct -- forced to map
// 		// TODO(geoah) WTF This is insanely hacky -- isn't it?
// 		if t := GetType(from); t != "" {
// 			m, err := Encode(data, false)
// 			if err != nil {
// 				return nil, err
// 			}
// 			m[attrCtx] = t
// 			for k, v := range m {
// 				if vt := GetType(reflect.TypeOf(v)); vt != "" {
// 					vm, err := Encode(v, true)
// 					if err != nil {
// 						return nil, err
// 					}
// 					vm[attrCtx] = vt
// 					m[k] = vm
// 				}
// 			}
// 			return m, nil
// 		}

// 		// decoding unknown struct to map
// 		if to.Kind() == reflect.Map {
// 			m, err := Encode(data, false)
// 			if err != nil {
// 				return nil, err
// 			}
// 			// fmt.Println("m", reflect.TypeOf(m), m)
// 			return m, nil
// 		}

// 		return data, nil
// 	}
// }

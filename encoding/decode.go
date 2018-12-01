package encoding

// import (
// 	"reflect"

// 	"github.com/mitchellh/mapstructure"
// )

// var (
// 	objectType = reflect.TypeOf(&Object{})
// )

// // Decode is a wrapper for mapstructure's Decode with our decodeHook that allows
// // decoding a map to the given structs
// // TODO move addCtx to an option
// // TODO move adding original block (@_) to an option
// func Decode(from map[string]interface{}, to interface{}, addCtx bool) error {
// 	dc := &mapstructure.DecoderConfig{
// 		Metadata:         &mapstructure.Metadata{},
// 		DecodeHook:       getDecodeHook(addCtx),
// 		Result:           to,
// 		TagName:          "json",
// 		WeaklyTypedInput: true,
// 	}
// 	dec, err := mapstructure.NewDecoder(dc)
// 	if err != nil {
// 		return err
// 	}

// 	if err := dec.Decode(from); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func getDecodeHook(addCtx bool) mapstructure.DecodeHookFuncType {
// 	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
// 		// fmt.Println("------------")
// 		// fmt.Println("from", from, from.Kind())
// 		// fmt.Println("to", to, to.Kind())
// 		// fmt.Println("data",data)

// 		if !addCtx {
// 			// TODO(geoah) explain this hack
// 			addCtx = true
// 			return data, nil
// 		}

// 		if to == objectType &&
// 			from == objectType &&
// 			reflect.TypeOf(data) == objectType {
// 			return data.(*Object).Map(), nil
// 		}

// 		// decoding map to struct
// 		if m, ok := data.(map[string]interface{}); ok {
// 			if to.Kind() == reflect.Map {
// 				return data, nil
// 			}

// 			t, ok := m[attrCtx].(string)
// 			if !ok {
// 				return data, nil
// 			}

// 			if to == reflect.TypeOf(&Object{}) {
// 				o := NewObject(m)
// 				return o, nil
// 			}

// 			v := GetInstance(t)
// 			if v == nil {
// 				return data, nil
// 			}

// 			delete(m, attrCtx)
// 			if err := Decode(m, v, true); err != nil {
// 				return nil, err
// 			}

// 			return v, nil
// 		}

// 		return data, nil
// 	}
// }

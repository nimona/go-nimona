package encoding

import (
	"errors"
	"fmt"
	"reflect"
)

// UntypeMap checks the type hints are correct, and removes them from the keys
func UntypeMap(m map[string]interface{}) (map[string]interface{}, error) {
	out := map[string]interface{}{}
	for k, v := range m {
		t := reflect.TypeOf(v)
		h := GetHintFromType(v)
		if h == "" {
			panic(fmt.Sprintf("untype: unsupported type k=%s t=%s v=%#v", k, t.String(), v))
		}
		eh := getFullType(k)
		if eh == "" {
			// key doesn't have type
			// TODO(geoah) is this even allowed?
			out[k] = v
			continue
		}
		// TODO(geoah) fix type checks
		// given []int{1, 2, 3} this will correctly hint A<i>
		// given []interface{}{1, 2, 3} this will correctly hint A<i>
		// given []interface{}{} this will __incorrectly__ hint A<?> and the type check will fail
		if h != eh {
			return nil, fmt.Errorf("untype: type hinted as %s, but is %s", eh, h)
		}
		// TODO should we be using type checks here?
		switch h {
		case HintMap:
			m, ok := v.(map[string]interface{})
			if !ok {
				return nil, errors.New("untype only supports map[string]interface{} maps")
			}
			var err error
			v, err = UntypeMap(m)
			if err != nil {
				return nil, err
			}
		case HintArray + "<" + HintMap + ">":
			vs, ok := v.([]interface{})
			if !ok {
				return nil, errors.New("untype only supports []interface{} for A<O>")
			}
			ovs := []interface{}{}
			for _, v := range vs {
				m, ok := v.(map[string]interface{})
				if !ok {
					return nil, errors.New("untype only supports map[string]interface{} maps")
				}
				ov, err := UntypeMap(m)
				if err != nil {
					return nil, err
				}
				ovs = append(ovs, ov)
			}
			v = ovs
		}
		k = k[:len(k)-len(h)-1]
		out[k] = v
	}
	return out, nil
}

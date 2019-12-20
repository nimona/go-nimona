package object

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"strings"

	"nimona.io/pkg/errors"
)

// Normalize maps to get them ready to be used as objects.
// This is supposed to convert a map's values into something more usable by
// using the type hints from the key as guide.
// A lot of what this does is due to the types go's JSON unmarshaller uses when
// unmarshalling into an `interface{}`.
// More info about this here: https://golang.org/pkg/encoding/json/#Unmarshal
//
// For example:
// * `"some-data:d": []float64{1, 2}` becomes `"some-data:d": []byte{1, 2}`
// * `"some-data:d": "AQI="` becomes `"some-data:d": []byte{1, 2}`
// * `"some-int:i": float64(7)` becomes `"some-int:i": uint64(7)`
// * `"some-int:i": "7"` becomes `"some-int:i": uint64(7)`
//
// NOTE: This should work for the most part but needs additional testing.
func Normalize(i interface{}) (map[string]interface{}, error) {
	return normalizeObject(i)
}

func normalizeFromKey(k string, i interface{}) (interface{}, error) {
	if i == nil {
		return nil, nil
	}
	ps := strings.Split(k, ":")
	t := ps[len(ps)-1]
	switch t[0] {
	case 'b':
		return normalizeBool(i)
	case 'a':
		a, ok := i.([]interface{})
		if !ok {
			return nil, errors.New("invalid array type, got " +
				reflect.TypeOf(i).String(),
			)
		}
		na := make([]interface{}, len(a))
		for i, v := range a {
			var nv interface{}
			var err error
			switch t[1] {
			case 'b':
				nv, err = normalizeBool(v)
			case 'o':
				nv, err = normalizeObject(v)
			case 's':
				nv, err = normalizeString(v)
			case 'd':
				nv, err = NormalizeData(v)
			case 'u':
				nv, err = normalizeUint(v)
			case 'i':
				nv, err = normalizeInt(v)
			case 'f':
				nv, err = normalizeFloat(v)
			default:
				return nil, errors.New("unknown array hint " + t)
			}
			if err != nil {
				return nil, errors.Wrap(err,
					errors.New("error normalising array value of "+t),
				)
			}
			na[i] = nv
		}
		return na, nil
	case 'o':
		return normalizeObject(i)
	case 's':
		return normalizeString(i)
	case 'd':
		return NormalizeData(i)
	case 'u':
		return normalizeUint(i)
	case 'i':
		return normalizeInt(i)
	case 'f':
		return normalizeFloat(i)
	}
	return nil, errors.New("unknown key hint " + t)
}

func normalizeBool(i interface{}) (bool, error) {
	switch v := i.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	}
	return false, errors.New("invalid bool type, got " +
		reflect.TypeOf(i).String(),
	)
}

func normalizeString(i interface{}) (string, error) {
	v, ok := i.(string)
	if !ok {
		return "", errors.New("invalid string type, got " +
			reflect.TypeOf(i).String(),
		)
	}
	return v, nil
}

func NormalizeData(i interface{}) ([]byte, error) {
	switch v := i.(type) {
	case []byte:
		return v, nil
	case []interface{}:
		d := make([]byte, len(v))
		for i, n := range v {
			u, err := normalizeUint(n)
			if err != nil {
				return nil, errors.Wrap(err, errors.New("could not normalize data"))
			}
			d[i] = uint8(u)
		}
		return d, nil
	case []float64:
		d := make([]byte, len(v))
		for i, n := range v {
			d[i] = uint8(n)
		}
		return d, nil
	case string:
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, errors.Wrap(err,
				errors.New("error decoding base64 data"),
			)
		}
		return b, nil
	}
	return nil, errors.New("unknown data type")
}

func normalizeUint(i interface{}) (uint64, error) {
	switch v := i.(type) {
	case float32:
		return uint64(v), nil
	case float64:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case int8:
		return uint64(v), nil
	case int16:
		return uint64(v), nil
	case int32:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return uint64(v), nil
	case string:
		nv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, errors.New("error parsing uint"))
		}
		return uint64(nv), nil
	}
	return 0, errors.New("invalid uint type")
}

func normalizeInt(i interface{}) (int64, error) {
	switch v := i.(type) {
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case string:
		nv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, errors.New("error parsing int"))
		}
		return nv, nil
	}
	return 0, errors.New("invalid int type")
}

func normalizeFloat(i interface{}) (float64, error) {
	switch v := i.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		nv, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, errors.Wrap(err, errors.New("error parsing float"))
		}
		return nv, nil
	}
	return 0, errors.New("invalid float type")
}

func normalizeObject(i interface{}) (map[string]interface{}, error) {
	nm := map[string]interface{}{}
	switch m := i.(type) {
	case map[string]interface{}:
		for k, v := range m {
			nv, err := normalizeFromKey(k, v)
			if err != nil {
				return nil, errors.Wrap(err,
					errors.New("error normalising value for map with key "+k),
				)
			}
			nm[k] = nv
		}
		return nm, nil
	}
	return nil, errors.New("unknown object type, got " +
		reflect.TypeOf(i).String(),
	)
}

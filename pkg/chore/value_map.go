package chore

import (
	"encoding/base64"
	"encoding/json"
	"sort"

	"github.com/buger/jsonparser"
	"nimona.io/pkg/errors"
)

func (v Map) Hint() Hint {
	return MapHint
}

func (v Map) _isValue() {
}

func (v Map) Hash() Hash {
	h := []byte{}
	ks := []string{}
	for k := range v {
		if len(k) > 0 && k[0] == '_' {
			continue
		}
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		iv := v[k]
		if iv == nil {
			continue
		}
		ivh := iv.Hash()
		if ivh.IsEmpty() {
			continue
		}

		k = k + ":"
		if _, ok := iv.(Map); ok {
			k += string(HashHint)
		} else {
			k += string(iv.Hint())
		}
		ikh := hashFromBytes(
			append(
				[]byte(StringHint),
				[]byte(k)...,
			),
		)
		h = append(
			h,
			ikh...,
		)
		h = append(
			h,
			ivh...,
		)
	}
	if len(h) == 0 {
		return EmptyHash
	}

	return hashFromBytes(h)
}

func jsonUnmarshalValue(
	h Hint,
	value []byte,
) (Value, error) {
	switch h {
	case BoolHint:
		var iv Bool
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case DataHint:
		iv, err := base64.StdEncoding.DecodeString(string(value))
		if err != nil {
			return nil, err
		}
		return Data(iv), nil
	case FloatHint:
		var iv Float
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case IntHint:
		var iv Int
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case MapHint:
		var iv Map = Map{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case StringHint:
		return String(value), nil
	case UintHint:
		var iv Uint
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case HashHint:
		return Hash(value), nil
	case BoolArrayHint:
		var iv BoolArray = BoolArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case DataArrayHint:
		var iv DataArray = DataArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case FloatArrayHint:
		var iv FloatArray = FloatArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case IntArrayHint:
		var iv IntArray = IntArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case MapArrayHint:
		var iv MapArray = MapArray{}
		if _, err := jsonparser.ArrayEach(value, func(
			value []byte,
			dataType jsonparser.ValueType,
			offset int,
			err error,
		) {
			iv = append(iv, Map{})
		}); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case StringArrayHint:
		var iv StringArray = StringArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case UintArrayHint:
		var iv UintArray = UintArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case HashArrayHint:
		var iv HashArray = HashArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	}
	return nil, errors.Error("map includes unimplemented hint")
}

func (v Map) UnmarshalJSON(b []byte) error {
	h := func(
		key []byte,
		value []byte,
		dataType jsonparser.ValueType,
		offset int,
	) error {
		k, h, err := ExtractHint(string(key))
		if err != nil {
			return err
		}
		if dataType == jsonparser.Null {
			return nil
		}
		iv, err := jsonUnmarshalValue(h, value)
		if err != nil {
			return err
		}
		v[k] = iv
		return nil
	}
	if err := jsonparser.ObjectEach(b, h); err != nil {
		return err
	}
	return nil
}

func (v Map) MarshalJSON() ([]byte, error) {
	m := map[string]Value{}
	for ik, iv := range v {
		if iv == nil {
			continue
		}
		switch ivv := iv.(type) {
		case Map:
			if len(ivv) == 0 {
				continue
			}
		case ArrayValue:
			if ivv.Len() == 0 {
				continue
			}
		}
		m[ik+":"+string(iv.Hint())] = iv
	}
	return json.Marshal(m)
}

package value

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/buger/jsonparser"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/object/hint"
)

type (
	Value interface {
		Hint() hint.Hint
		_isValue()
	}

	// TODO switch Range's stop?
	ArrayValue interface {
		Value
		Len() int
		Range(func(int, Value) (stop bool))
		_isArray()
	}

	// basic types
	Bool   bool
	Data   []byte
	Float  float64
	Int    int64
	Map    map[string]Value
	String string
	Uint   uint64
	CID    string

	// array types
	BoolArray   []Bool
	DataArray   []Data
	FloatArray  []Float
	IntArray    []Int
	MapArray    []Map
	StringArray []String
	UintArray   []Uint
	CIDArray    []CID
)

// TODO(geoah) do we need this?
func (v CID) IsEmpty() bool {
	return string(v) == ""
}

func (v CID) String() string {
	return string(v)
}

func (v Bool) Hint() hint.Hint { return hint.Bool }
func (v Bool) _isValue()       {}

func (v *Bool) UnmarshalJSON(b []byte) error {
	var iv bool
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Bool(iv)
	return nil
}

func (v Data) Hint() hint.Hint { return hint.Data }
func (v Data) _isValue()       {}

func (v *Data) UnmarshalJSON(b []byte) error {
	iv := []byte{}
	err := json.Unmarshal(b, &iv)
	if err != nil {
		fmt.Println(err)
		return err
	}
	*v = Data(iv)
	return nil
}

func (v Float) Hint() hint.Hint { return hint.Float }
func (v Float) _isValue()       {}

func (v *Float) UnmarshalJSON(b []byte) error {
	var iv float64
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Float(iv)
	return nil
}

func (v Int) Hint() hint.Hint { return hint.Int }
func (v Int) _isValue()       {}

func (v *Int) UnmarshalJSON(b []byte) error {
	var iv int64
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Int(iv)
	return nil
}

func (v Map) Hint() hint.Hint { return hint.Map }
func (v Map) _isValue()       {}

func jsonUnmarshalValue(
	h hint.Hint,
	value []byte,
) (Value, error) {
	switch h {
	case hint.Bool:
		var iv Bool
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case hint.Data:
		iv, err := base64.StdEncoding.DecodeString(string(value))
		if err != nil {
			fmt.Println(string(value))
			return nil, err
		}
		return Data(iv), nil
	case hint.Float:
		var iv Float
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case hint.Int:
		var iv Int
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case hint.Map:
		var iv Map = Map{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case hint.String:
		return String(string(value)), nil
	case hint.Uint:
		var iv Uint
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case hint.CID:
		return CID(value), nil
	case hint.BoolArray:
		var iv BoolArray = BoolArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case hint.DataArray:
		var iv DataArray = DataArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case hint.FloatArray:
		var iv FloatArray = FloatArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case hint.IntArray:
		var iv IntArray = IntArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case hint.MapArray:
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
	case hint.StringArray:
		var iv StringArray = StringArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case hint.UintArray:
		var iv UintArray = UintArray{}
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case hint.CIDArray:
		var iv CIDArray = CIDArray{}
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
		k, h, err := hint.Extract(string(key))
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

func (v String) Hint() hint.Hint { return hint.String }
func (v String) _isValue()       {}

func (v Uint) Hint() hint.Hint { return hint.Uint }
func (v Uint) _isValue()       {}

func (v CID) Hint() hint.Hint { return hint.CID }
func (v CID) _isValue()       {}

func (v *Uint) UnmarshalJSON(b []byte) error {
	var iv uint64
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Uint(iv)
	return nil
}

func (v BoolArray) Hint() hint.Hint { return hint.BoolArray }
func (v BoolArray) _isValue()       {}
func (v BoolArray) _isArray()       {}
func (v BoolArray) Len() int        { return len(v) }
func (v BoolArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v DataArray) Hint() hint.Hint { return hint.DataArray }
func (v DataArray) _isValue()       {}
func (v DataArray) _isArray()       {}
func (v DataArray) Len() int        { return len(v) }
func (v DataArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v FloatArray) Hint() hint.Hint { return hint.FloatArray }
func (v FloatArray) _isValue()       {}
func (v FloatArray) _isArray()       {}
func (v FloatArray) Len() int        { return len(v) }
func (v FloatArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v IntArray) Hint() hint.Hint { return hint.IntArray }
func (v IntArray) _isValue()       {}
func (v IntArray) _isArray()       {}
func (v IntArray) Len() int        { return len(v) }
func (v IntArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v MapArray) Hint() hint.Hint { return hint.MapArray }
func (v MapArray) _isValue()       {}
func (v MapArray) _isArray()       {}
func (v MapArray) Len() int        { return len(v) }
func (v MapArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v StringArray) Hint() hint.Hint { return hint.StringArray }
func (v StringArray) _isValue()       {}
func (v StringArray) _isArray()       {}
func (v StringArray) Len() int        { return len(v) }
func (v StringArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v UintArray) Hint() hint.Hint { return hint.UintArray }
func (v UintArray) _isValue()       {}
func (v UintArray) _isArray()       {}
func (v UintArray) Len() int        { return len(v) }
func (v UintArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v CIDArray) Hint() hint.Hint { return hint.CIDArray }
func (v CIDArray) _isValue()       {}
func (v CIDArray) _isArray()       {}
func (v CIDArray) Len() int        { return len(v) }
func (v CIDArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

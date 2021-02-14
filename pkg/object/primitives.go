package object

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/buger/jsonparser"

	"nimona.io/pkg/errors"
)

type (
	Hint  string
	Value interface {
		Hint() Hint
		_isValue()
	}
	ArrayValue interface {
		Value
		Len() int
		Range(func(int, Value) bool)
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
	Hash   string
	// array types
	BoolArray   []Bool
	DataArray   []Data
	FloatArray  []Float
	IntArray    []Int
	MapArray    []Map
	ObjectArray []*Object // nolint: golint
	StringArray []String
	UintArray   []Uint
	HashArray   []Hash
)

const (
	// basic hints
	BoolHint   Hint = "b"
	DataHint   Hint = "d"
	FloatHint  Hint = "f"
	IntHint    Hint = "i"
	MapHint    Hint = "m"
	ObjectHint Hint = "o"
	StringHint Hint = "s"
	UintHint   Hint = "u"
	HashHint   Hint = "r"
	// array hints
	BoolArrayHint   Hint = "ab"
	DataArrayHint   Hint = "ad"
	FloatArrayHint  Hint = "af"
	IntArrayHint    Hint = "ai"
	MapArrayHint    Hint = "am"
	ObjectArrayHint Hint = "ao"
	StringArrayHint Hint = "as"
	UintArrayHint   Hint = "au"
	HashArrayHint   Hint = "ar"
)

var hints = map[string]Hint{
	// basic hints
	string(BoolHint):   BoolHint,
	string(DataHint):   DataHint,
	string(FloatHint):  FloatHint,
	string(IntHint):    IntHint,
	string(MapHint):    MapHint,
	string(ObjectHint): ObjectHint,
	string(StringHint): StringHint,
	string(UintHint):   UintHint,
	string(HashHint):   HashHint,
	// array hints
	string(BoolArrayHint):   BoolArrayHint,
	string(DataArrayHint):   DataArrayHint,
	string(FloatArrayHint):  FloatArrayHint,
	string(IntArrayHint):    IntArrayHint,
	string(MapArrayHint):    MapArrayHint,
	string(ObjectArrayHint): ObjectArrayHint,
	string(StringArrayHint): StringArrayHint,
	string(UintArrayHint):   UintArrayHint,
	string(HashArrayHint):   HashArrayHint,
}

func splitHint(b []byte) (string, Hint, error) {
	ps := strings.Split(string(b), ":")
	if len(ps) != 2 {
		return "", "", errors.New("invalid hinted key")
	}
	h, ok := hints[ps[1]]
	if !ok {
		return "", "", errors.New("unknown hint")
	}
	return ps[0], h, nil
}

func (v Bool) Hint() Hint { return BoolHint }
func (v Bool) _isValue()  {}

func (v *Bool) UnmarshalJSON(b []byte) error {
	var iv bool
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Bool(iv)
	return nil
}

func (v Data) Hint() Hint { return DataHint }
func (v Data) _isValue()  {}

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

func (v Float) Hint() Hint { return FloatHint }
func (v Float) _isValue()  {}

func (v *Float) UnmarshalJSON(b []byte) error {
	var iv float64
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Float(iv)
	return nil
}

func (v Int) Hint() Hint { return IntHint }
func (v Int) _isValue()  {}

func (v *Int) UnmarshalJSON(b []byte) error {
	var iv int64
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Int(iv)
	return nil
}

func (v Map) Hint() Hint { return MapHint }
func (v Map) _isValue()  {}

func jsonUnmarshalValue(
	hint Hint,
	value []byte,
) (Value, error) {
	if len(hints) == 0 {
		return nil, errors.New("no hints supplied")
	}
	switch hint {
	case BoolHint:
		var iv Bool
		if err := json.Unmarshal(value, &iv); err != nil {
			return nil, err
		}
		return iv, nil
	case DataHint:
		iv, err := base64.StdEncoding.DecodeString(string(value))
		if err != nil {
			fmt.Println(string(value))
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
	case ObjectHint:
		var iv *Object = &Object{}
		if err := json.Unmarshal(value, iv); err != nil {
			return nil, err
		}
		return iv, nil
	case StringHint:
		return String(string(value)), nil
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
	case ObjectArrayHint:
		var iv ObjectArray = ObjectArray{}
		if _, err := jsonparser.ArrayEach(value, func(
			value []byte,
			dataType jsonparser.ValueType,
			offset int,
			err error,
		) {
			iv = append(iv, &Object{})
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
	return nil, errors.New("map includes unimplemented hint")
}

func (v Map) UnmarshalJSON(b []byte) error {
	h := func(
		key []byte,
		value []byte,
		dataType jsonparser.ValueType,
		offset int,
	) error {
		k, h, err := splitHint(key)
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

func (v Object) Hint() Hint { return ObjectHint }
func (v Object) _isValue()  {}

func (v String) Hint() Hint { return StringHint }
func (v String) _isValue()  {}

func (v Uint) Hint() Hint { return UintHint }
func (v Uint) _isValue()  {}

func (v Hash) Hint() Hint { return HashHint }
func (v Hash) _isValue()  {}

func (v *Uint) UnmarshalJSON(b []byte) error {
	var iv uint64
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Uint(iv)
	return nil
}

func (v BoolArray) Hint() Hint { return BoolArrayHint }
func (v BoolArray) _isValue()  {}
func (v BoolArray) _isArray()  {}
func (v BoolArray) Len() int   { return len(v) }
func (v BoolArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v DataArray) Hint() Hint { return DataArrayHint }
func (v DataArray) _isValue()  {}
func (v DataArray) _isArray()  {}
func (v DataArray) Len() int   { return len(v) }
func (v DataArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v FloatArray) Hint() Hint { return FloatArrayHint }
func (v FloatArray) _isValue()  {}
func (v FloatArray) _isArray()  {}
func (v FloatArray) Len() int   { return len(v) }
func (v FloatArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v IntArray) Hint() Hint { return IntArrayHint }
func (v IntArray) _isValue()  {}
func (v IntArray) _isArray()  {}
func (v IntArray) Len() int   { return len(v) }
func (v IntArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v MapArray) Hint() Hint { return MapArrayHint }
func (v MapArray) _isValue()  {}
func (v MapArray) _isArray()  {}
func (v MapArray) Len() int   { return len(v) }
func (v MapArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v ObjectArray) Hint() Hint { return ObjectArrayHint }
func (v ObjectArray) _isValue()  {}
func (v ObjectArray) _isArray()  {}
func (v ObjectArray) Len() int   { return len(v) }
func (v ObjectArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v StringArray) Hint() Hint { return StringArrayHint }
func (v StringArray) _isValue()  {}
func (v StringArray) _isArray()  {}
func (v StringArray) Len() int   { return len(v) }
func (v StringArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v UintArray) Hint() Hint { return UintArrayHint }
func (v UintArray) _isValue()  {}
func (v UintArray) _isArray()  {}
func (v UintArray) Len() int   { return len(v) }
func (v UintArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v HashArray) Hint() Hint { return HashArrayHint }
func (v HashArray) _isValue()  {}
func (v HashArray) _isArray()  {}
func (v HashArray) Len() int   { return len(v) }
func (v HashArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

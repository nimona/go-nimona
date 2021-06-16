package chore

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/buger/jsonparser"

	"nimona.io/internal/encoding/base58"
	"nimona.io/pkg/errors"
)

const (
	EmptyHash Hash = ""
)

type (
	Value interface {
		Hash() Hash
		Hint() Hint
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
	Hash   string

	// array types
	BoolArray   []Bool
	DataArray   []Data
	FloatArray  []Float
	IntArray    []Int
	MapArray    []Map
	StringArray []String
	UintArray   []Uint
	HashArray   []Hash
)

func hashFromBytes(d []byte) Hash {
	if d == nil {
		return EmptyHash
	}
	b := sha256.Sum256(d)
	return Hash(base58.Encode(b[:]))
}

func (v Hash) Bytes() ([]byte, error) {
	return base58.Decode(string(v))
}

func (v Hash) IsEmpty() bool {
	return string(v) == ""
}

func (v Hash) Equal(h Hash) bool {
	return h == v
}

func (v Hash) String() string {
	return string(v)
}

func (v Bool) Hint() Hint { return BoolHint }
func (v Bool) _isValue()  {}

func (v Bool) Hash() Hash {
	if !v {
		return hashFromBytes([]byte{0})
	}
	return hashFromBytes([]byte{1})
}

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

func (v Data) Hash() Hash {
	return hashFromBytes(v)
}

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

func (v Float) Hash() Hash {
	// replacing ben's implementation with something less custom, based on:
	// * https://github.com/benlaurie/objecthash
	// * https://play.golang.org/p/3xraud43pi
	// examples of same results in other languages
	// * ruby: `[7.30363941192626953125].pack('G').unpack('B*').first`
	// * js: `http://weitz.de/ieee`
	//
	// NOTE(geoah): I have removed the inf and nan hashing for now,
	// we can revisit them once we better understand their usecases.
	switch {
	case math.IsInf(float64(v), 1):
		panic(errors.Error("float inf is not currently supported"))
	case math.IsInf(float64(v), -1):
		panic(errors.Error("float -inf is not currently supported"))
	case math.IsNaN(float64(v)):
		panic(errors.Error("float nan is not currently supported"))
	default:
		return hashFromBytes(
			[]byte(
				fmt.Sprintf(
					"%d",
					math.Float64bits(float64(v)),
				),
			),
		)
	}
}

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

func (v Int) Hash() Hash {
	return hashFromBytes(
		[]byte(
			fmt.Sprintf(
				"%d",
				int64(v),
			),
		),
	)
}

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

func (v String) Hint() Hint { return StringHint }
func (v String) _isValue()  {}

func (v String) Hash() Hash {
	if string(v) == "" {
		return EmptyHash
	}
	return hashFromBytes(
		[]byte(string(v)),
	)
}

func (v Uint) Hint() Hint { return UintHint }
func (v Uint) _isValue()  {}

func (v Uint) Hash() Hash {
	return hashFromBytes(
		[]byte(
			fmt.Sprintf(
				"%d",
				uint64(v),
			),
		),
	)
}

func (v Hash) Hint() Hint { return HashHint }
func (v Hash) _isValue()  {}

func (v Hash) Hash() Hash {
	return v
}

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

func (v BoolArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}
func (v BoolArray) _isArray() {}
func (v BoolArray) Len() int  { return len(v) }
func (v BoolArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v DataArray) Hint() Hint { return DataArrayHint }
func (v DataArray) _isValue()  {}

func (v DataArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}
func (v DataArray) _isArray() {}
func (v DataArray) Len() int  { return len(v) }
func (v DataArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v FloatArray) Hint() Hint { return FloatArrayHint }
func (v FloatArray) _isValue()  {}

func (v FloatArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}
func (v FloatArray) _isArray() {}
func (v FloatArray) Len() int  { return len(v) }
func (v FloatArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v IntArray) Hint() Hint { return IntArrayHint }
func (v IntArray) _isValue()  {}

func (v IntArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}
func (v IntArray) _isArray() {}
func (v IntArray) Len() int  { return len(v) }
func (v IntArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v MapArray) Hint() Hint { return MapArrayHint }
func (v MapArray) _isValue()  {}

func (v MapArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}
func (v MapArray) _isArray() {}
func (v MapArray) Len() int  { return len(v) }
func (v MapArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v StringArray) Hint() Hint { return StringArrayHint }
func (v StringArray) _isValue()  {}

func (v StringArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}
func (v StringArray) _isArray() {}
func (v StringArray) Len() int  { return len(v) }
func (v StringArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v UintArray) Hint() Hint { return UintArrayHint }
func (v UintArray) _isValue()  {}

func (v UintArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}
func (v UintArray) _isArray() {}
func (v UintArray) Len() int  { return len(v) }
func (v UintArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

func (v HashArray) Hint() Hint { return HashArrayHint }
func (v HashArray) _isValue()  {}

func (v HashArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}
func (v HashArray) _isArray() {}
func (v HashArray) Len() int  { return len(v) }
func (v HashArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

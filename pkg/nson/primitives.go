package nson

type (
	Value interface {
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

func (v Bool) Hint() Hint { return BoolHint }
func (v Bool) _isValue()  {}

func (v Data) Hint() Hint { return DataHint }
func (v Data) _isValue()  {}

func (v Float) Hint() Hint { return FloatHint }
func (v Float) _isValue()  {}

func (v Int) Hint() Hint { return IntHint }
func (v Int) _isValue()  {}

func (v Map) Hint() Hint { return MapHint }
func (v Map) _isValue()  {}

func (v String) Hint() Hint { return StringHint }
func (v String) _isValue()  {}

func (v Uint) Hint() Hint { return UintHint }
func (v Uint) _isValue()  {}

func (v CID) Hint() Hint { return CIDHint }
func (v CID) _isValue()  {}

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

func (v CIDArray) Hint() Hint { return CIDArrayHint }
func (v CIDArray) _isValue()  {}
func (v CIDArray) _isArray()  {}
func (v CIDArray) Len() int   { return len(v) }
func (v CIDArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

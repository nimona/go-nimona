package chore

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

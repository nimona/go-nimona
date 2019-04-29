package immutable

type (
	typeHinted interface {
		typeHint() string
		primitive() interface{}
	}
	Value struct {
		kind typeHinted
	}
	boolValue   struct{ value bool }
	stringValue struct{ value string }
	intValue    struct{ value int64 }
	floatValue  struct{ value float64 }
	bytesValue  struct{ value []byte }
	mapValue    struct{ value Map }
	// listValue   struct{ value List }
)

func (v Value) BoolValue() bool {
	if x, ok := v.kind.(boolValue); ok {
		return x.value
	}
	return false
}
func (v Value) StringValue() string {
	if x, ok := v.kind.(stringValue); ok {
		return x.value
	}
	return ""
}

func (v Value) IntValue() int64 {
	if x, ok := v.kind.(intValue); ok {
		return x.value
	}
	return 0
}

func (v Value) FloatValue() float64 {
	if x, ok := v.kind.(floatValue); ok {
		return x.value
	}
	return 0
}

func (v Value) BytesValue() []byte {
	if x, ok := v.kind.(bytesValue); ok {
		return x.value
	}
	return nil
}

func (v Value) MapValue() Map {
	if x, ok := v.kind.(mapValue); ok {
		return x.value
	}
	return Map{}
}

// func (v Value) ListValue() List {
// 	if x, ok := v.kind.(listValue); ok {
// 		return x.value
// 	}
// 	return List{}
// }

func (v Value) Primitive() interface{} {
	switch b := v.kind.(type) {
	case boolValue:
		return v.kind.primitive()
	case stringValue:
		return v.kind.primitive()
	case intValue:
		return v.kind.primitive()
	case floatValue:
		return v.kind.primitive()
	case bytesValue:
		return v.kind.primitive()
	case mapValue:
		return b.value.Primitive()
		// case listValue:
		// 	return v.kind.primitive()
	}
	return nil
}

func (v Value) PrimitiveHinted() interface{} {
	switch b := v.kind.(type) {
	case boolValue:
		return v.kind.primitive()
	case stringValue:
		return v.kind.primitive()
	case intValue:
		return v.kind.primitive()
	case floatValue:
		return v.kind.primitive()
	case bytesValue:
		return v.kind.primitive()
	case mapValue:
		return b.value.PrimitiveHinted()
		// case listValue:
		// 	return v.kind.primitive()
	}
	return nil
}

const (
	boolTypeHint   = "b"
	stringTypeHint = "s"
	intTypeHint    = "i"
	floatTypeHint  = "f"
	bytesTypeHint  = "d"
	mapTypeHint    = "o"
	listTypeHint   = "a"
)

func (v boolValue) typeHint() string   { return boolTypeHint }
func (v stringValue) typeHint() string { return stringTypeHint }
func (v intValue) typeHint() string    { return intTypeHint }
func (v floatValue) typeHint() string  { return floatTypeHint }
func (v bytesValue) typeHint() string  { return bytesTypeHint }
func (v mapValue) typeHint() string    { return mapTypeHint }

// func (v listValue) typeHint() string   { return listTypeHint }

func (v boolValue) primitive() interface{}   { return v.value }
func (v stringValue) primitive() interface{} { return v.value }
func (v intValue) primitive() interface{}    { return v.value }
func (v floatValue) primitive() interface{}  { return v.value }
func (v bytesValue) primitive() interface{}  { return v.value }
func (v mapValue) primitive() interface{}    { return v.value }

// func (v listValue) primitive() interface{}   { return v.value }

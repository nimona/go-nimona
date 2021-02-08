package object

// type (
// 	// TypeHint are the hints of a member's type
// 	TypeHint string
// )

// // String implements the Stringer interface
// func (t TypeHint) String() string {
// 	return string(t)
// }

// const (
// 	HintUndefined TypeHint = ""
// 	HintArray     TypeHint = "a"
// 	HintBool      TypeHint = "b"
// 	HintData      TypeHint = "d"
// 	HintFloat     TypeHint = "f"
// 	HintInt       TypeHint = "i"
// 	HintMap       TypeHint = "m"
// 	HintNil       TypeHint = "n"
// 	HintObject    TypeHint = "o"
// 	HintRef       TypeHint = "r"
// 	HintString    TypeHint = "s"
// 	HintUint      TypeHint = "u"
// )

// var (
// 	hints = map[string]TypeHint{
// 		"":  HintUndefined,
// 		"a": HintArray,
// 		"b": HintBool,
// 		"d": HintData,
// 		"f": HintFloat,
// 		"i": HintInt,
// 		"m": HintMap,
// 		"n": HintNil,
// 		"o": HintObject,
// 		"r": HintRef,
// 		"s": HintString,
// 		"u": HintUint,
// 	}
// )

// // GetTypeHint returns a TypeHint from a string
// func GetTypeHint(t string) TypeHint {
// 	h := HintUndefined
// 	if t, ok := hints[t]; ok {
// 		h = t
// 	}
// 	return h
// }

package hint

import (
	"strings"

	"nimona.io/pkg/errors"
)

type Hint string

const (
	// basic hints
	Bool   Hint = "b"
	Data   Hint = "d"
	Float  Hint = "f"
	Int    Hint = "i"
	Map    Hint = "m"
	String Hint = "s"
	Uint   Hint = "u"
	CID    Hint = "r"
	// array hints
	BoolArray   Hint = "ab"
	DataArray   Hint = "ad"
	FloatArray  Hint = "af"
	IntArray    Hint = "ai"
	MapArray    Hint = "am"
	ObjectArray Hint = "ao"
	StringArray Hint = "as"
	UintArray   Hint = "au"
	CIDArray    Hint = "ar"
)

var hints = map[string]Hint{
	// basic hints
	string(Bool):   Bool,
	string(Data):   Data,
	string(Float):  Float,
	string(Int):    Int,
	string(Map):    Map,
	string(String): String,
	string(Uint):   Uint,
	string(CID):    CID,
	// array hints
	string(BoolArray):   BoolArray,
	string(DataArray):   DataArray,
	string(FloatArray):  FloatArray,
	string(IntArray):    IntArray,
	string(MapArray):    MapArray,
	string(ObjectArray): ObjectArray,
	string(StringArray): StringArray,
	string(UintArray):   UintArray,
	string(CIDArray):    CIDArray,
}

func Extract(key string) (string, Hint, error) {
	ps := strings.Split(key, ":")
	if len(ps) != 2 {
		return "", "", errors.Error("extractHint: invalid hinted key " + key)
	}
	h, ok := hints[ps[1]]
	if !ok {
		return "", "", errors.Error("extractHint: invalid hint " + ps[1])
	}
	return ps[0], h, nil
}

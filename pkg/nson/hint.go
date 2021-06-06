package nson

import (
	"strings"

	"nimona.io/pkg/errors"
)

type Hint string

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
	CIDHint    Hint = "r"
	// array hints
	BoolArrayHint   Hint = "ab"
	DataArrayHint   Hint = "ad"
	FloatArrayHint  Hint = "af"
	IntArrayHint    Hint = "ai"
	MapArrayHint    Hint = "am"
	ObjectArrayHint Hint = "ao"
	StringArrayHint Hint = "as"
	UintArrayHint   Hint = "au"
	CIDArrayHint    Hint = "ar"
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
	string(CIDHint):    CIDHint,
	// array hints
	string(BoolArrayHint):   BoolArrayHint,
	string(DataArrayHint):   DataArrayHint,
	string(FloatArrayHint):  FloatArrayHint,
	string(IntArrayHint):    IntArrayHint,
	string(MapArrayHint):    MapArrayHint,
	string(ObjectArrayHint): ObjectArrayHint,
	string(StringArrayHint): StringArrayHint,
	string(UintArrayHint):   UintArrayHint,
	string(CIDArrayHint):    CIDArrayHint,
}

func ExtractHint(key string) (string, Hint, error) {
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

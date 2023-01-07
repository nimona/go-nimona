package nimona

type CborFixture struct {
	String         string       `cborgen:"string,omitempty" json:"string,omitempty"`
	Uint64         uint64       `cborgen:"uint64,omitempty" json:"uint64,omitempty"`
	Int64          int64        `cborgen:"int64,omitempty" json:"int64,omitempty"`
	Bytes          []byte       `cborgen:"bytes,omitempty" json:"bytes,omitempty"`
	Bool           bool         `cborgen:"bool,omitempty" json:"bool,omitempty"`
	Map            *CborFixture `cborgen:"map,omitempty" json:"map,omitempty"`
	RepeatedString []string     `cborgen:"repeatedstring,omitempty" json:"repeatedstring,omitempty"`
	RepeatedUint64 []uint64     `cborgen:"repeateduint64,omitempty" json:"repeateduint64,omitempty"`
	RepeatedInt64  []int64      `cborgen:"repeatedint64,omitempty" json:"repeatedint64,omitempty"`
	RepeatedBytes  [][]byte     `cborgen:"repeatedbytes,omitempty" json:"repeatedbytes,omitempty"`
	// RepeatedBool   []bool `cborgen:"repeatedbool,omitempty" json:"repeatedbool,omitempty"`
	RepeatedMap     []*CborFixture `cborgen:"repeatedmap,omitempty" json:"repeatedmap,omitempty"`
	EphemeralString string         `cborgen:"_ephemeralString,omitempty" json:"_ephemeralString,omitempty"`
}

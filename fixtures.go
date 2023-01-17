package nimona

type CborFixture struct {
	String         string       `cborgen:"string,omitempty"`
	Uint64         uint64       `cborgen:"uint64,omitempty"`
	Int64          int64        `cborgen:"int64,omitempty"`
	Bytes          []byte       `cborgen:"bytes,omitempty"`
	Bool           bool         `cborgen:"bool,omitempty"`
	Map            *CborFixture `cborgen:"map,omitempty"`
	RepeatedString []string     `cborgen:"repeatedstring,omitempty"`
	RepeatedUint64 []uint64     `cborgen:"repeateduint64,omitempty"`
	RepeatedInt64  []int64      `cborgen:"repeatedint64,omitempty"`
	RepeatedBytes  [][]byte     `cborgen:"repeatedbytes,omitempty"`
	// RepeatedBool   []bool `cborgen:"repeatedbool,omitempty"`
	RepeatedMap     []*CborFixture `cborgen:"repeatedmap,omitempty"`
	EphemeralString string         `cborgen:"_ephemeralString,omitempty"`

	// Tuples
	DocumentID DocumentID `cborgen:"documentID,omitempty"`
}

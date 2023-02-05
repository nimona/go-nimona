package nimona

type CborFixture struct {
	Metadata       Metadata     `nimona:"$metadata,omitempty,type=test/fixture"`
	String         string       `nimona:"string,omitempty"`
	Uint64         uint64       `nimona:"uint64,omitempty"`
	Int64          int64        `nimona:"int64,omitempty"`
	Bytes          []byte       `nimona:"bytes,omitempty"`
	Bool           bool         `nimona:"bool,omitempty"`
	Map            *CborFixture `nimona:"map,omitempty"`
	RepeatedString []string     `nimona:"repeatedstring,omitempty"`
	RepeatedUint64 []uint64     `nimona:"repeateduint64,omitempty"`
	RepeatedInt64  []int64      `nimona:"repeatedint64,omitempty"`
	RepeatedBytes  [][]byte     `nimona:"repeatedbytes,omitempty"`
	// RepeatedBool    []bool         `nimona:"repeatedbool,omitempty"`
	RepeatedMap     []*CborFixture `nimona:"repeatedmap,omitempty"`
	EphemeralString string         `nimona:"_ephemeralString,omitempty"`
	DocumentID      DocumentID     `nimona:"documentID,omitempty"`
}

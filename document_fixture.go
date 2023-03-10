package nimona

// nolint: unused // used in tests
type documentFixture struct {
	Metadata       Metadata           `nimona:"$metadata,omitempty,type=test/fixture"`
	StringConst    string             `nimona:"stringConst,const=foo"`
	String         string             `nimona:"string,omitempty"`
	Uint64         uint64             `nimona:"uint64,omitempty"`
	Int64          int64              `nimona:"int64,omitempty"`
	Bytes          []byte             `nimona:"bytes,omitempty"`
	Bool           bool               `nimona:"bool,omitempty"`
	MapPtr         *documentFixture   `nimona:"mapPtr,omitempty"`
	RepeatedString []string           `nimona:"repeatedstring,omitempty"`
	RepeatedUint64 []uint64           `nimona:"repeateduint64,omitempty"`
	RepeatedInt64  []int64            `nimona:"repeatedint64,omitempty"`
	RepeatedBytes  [][]byte           `nimona:"repeatedbytes,omitempty"`
	RepeatedBool   []bool             `nimona:"repeatedbool,omitempty"`
	RepeatedMap    []documentFixture  `nimona:"repeatedmap,omitempty"`
	RepeatedMapPtr []*documentFixture `nimona:"repeatedmapPtr,omitempty"`
	// edge cases
	EphemeralString string     `nimona:"_ephemeralString,omitempty"`
	DocumentID      DocumentID `nimona:"documentID,omitempty"`
}

// nolint: unused // used in tests
type documentFixtureWithType struct {
	_      string `nimona:"$type,omitempty,type=foobar"`
	String string `nimona:"string,omitempty"`
}

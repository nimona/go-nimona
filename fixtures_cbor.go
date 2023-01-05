package nimona

type CborFixture struct {
	String         string       `cborgen:"omitempty"`
	Uint64         uint64       `cborgen:"omitempty"`
	Int64          int64        `cborgen:"omitempty"`
	Bytes          []byte       `cborgen:"omitempty"`
	Bool           bool         `cborgen:"omitempty"`
	Map            *CborFixture `cborgen:"omitempty"`
	RepeatedString []string     `cborgen:"omitempty"`
	RepeatedUint64 []uint64     `cborgen:"omitempty"`
	RepeatedInt64  []int64      `cborgen:"omitempty"`
	RepeatedBytes  [][]byte     `cborgen:"omitempty"`
	// RepeatedBool   []bool `cborgen:"omitempty"`
	RepeatedMap []*CborFixture `cborgen:"omitempty"`
}

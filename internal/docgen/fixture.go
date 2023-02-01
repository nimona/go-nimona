package docgen

type Fixture struct {
	StringConst    string     `nimona:"stringConst,omitempty,const=foo"`
	String         string     `nimona:"string,omitempty"`
	Uint64         uint64     `nimona:"uint64,omitempty"`
	Int64          int64      `nimona:"int64,omitempty"`
	Bytes          []byte     `nimona:"bytes,omitempty"`
	Bool           bool       `nimona:"bool,omitempty"`
	MapPtr         *Fixture   `nimona:"mapPtr,omitempty"`
	RepeatedString []string   `nimona:"repeatedstring,omitempty"`
	RepeatedUint64 []uint64   `nimona:"repeateduint64,omitempty"`
	RepeatedInt64  []int64    `nimona:"repeatedint64,omitempty"`
	RepeatedBytes  [][]byte   `nimona:"repeatedbytes,omitempty"`
	RepeatedBool   []bool     `nimona:"repeatedbool,omitempty"`
	RepeatedMap    []Fixture  `nimona:"repeatedmap,omitempty"`
	RepeatedMapPtr []*Fixture `nimona:"repeatedmapPtr,omitempty"`
}

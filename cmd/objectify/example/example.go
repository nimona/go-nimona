package example

import "nimona.io/go/encoding"

//go:generate go run nimona.io/go/cmd/objectify -schema test/inn -type InnerFoo -out inner_foo_generated.go
//go:generate go run nimona.io/go/cmd/objectify -schema test/foo -type Foo -out foo_generated.go

type InnerFoo struct {
	InnerBar      string      `fluffy:"inner_bar:s"`
	MoreInnerFoos []*InnerFoo `json:"inner_foos"`
	I             int
	I8            int8
	I16           int16
	I32           int32
	I64           int64
	U             uint
	U8            uint8
	U16           uint16
	U32           uint32
	F32           float32
	F64           float64
	Ai8           []int8
	Ai16          []int16
	Ai32          []int32
	Ai64          []int64
	Au16          []uint16
	Au32          []uint32
	Af32          []float32
	Af64          []float64
	// AAi           [][]int
	// AAf           [][]float32
	// O             map[string]interface{}
	// B             bool
}

type Foo struct {
	RawObject *encoding.Object `fluffy:"@"`
	Bar       string           `fluffy:"bar:s"`
	Bars      []string         `fluffys:"bars"`
	InnerFoo  *InnerFoo        `fluffy:"inner_foo:O,object"`
	InnerFoos []*InnerFoo      `fluffy:"inner_foos:A<O>"`
}

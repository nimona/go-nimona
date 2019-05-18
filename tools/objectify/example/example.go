package example

import "nimona.io/pkg/object"

// nolint
//go:generate $GOBIN/objectify -schema test/inn -type InnerFoo -in example.go -out inner_foo_generated.go
//go:generate $GOBIN/objectify -schema test/foo -type Foo -in example.go -out foo_generated.go

// InnerFoo -
// nolint
type InnerFoo struct {
	InnerBar      string      `json:"inner_bar"`
	InnerBars     []string    `json:"inner_bars"`
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
	AAi           [][]int
	AAf           [][]float32
	AAs           [][]string
	// O             map[string]interface{}
	B bool
}

// Foo -
type Foo struct {
	// RawObject *object.Object   `json:"@"`
	Bar       string           `json:"bar"`
	Bars      []string         `json:"bars"`
	InnerFoo  *InnerFoo        `json:"inner_foo"`
	InnerFoos []*InnerFoo      `json:"inner_foos"`
	Object    *object.Object   `json:"object"`
	Objects   []*object.Object `json:"objects"`
}

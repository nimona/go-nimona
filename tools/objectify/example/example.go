package example

import "nimona.io/pkg/object"

// nolint
//go:generate $GOBIN/objectify -schema test/inn -type InnerFoo -in example.go -out inner_foo_generated.go
//go:generate $GOBIN/objectify -schema test/foo -type Foo -in example.go -out foo_generated.go

// InnerFoo -
// nolint
type InnerFoo struct {
	InnerBar      string      `json:"inner_bar:s"`
	InnerBars     []string    `json:"inner_bars:as"`
	MoreInnerFoos []*InnerFoo `json:"inner_foos:ao"`
	I             int         `json:"I:i"`
	I8            int8        `json:"I8:i"`
	I16           int16       `json:"I16:i"`
	I32           int32       `json:"I32:i"`
	I64           int64       `json:"I64:i"`
	U             uint        `json:"U:u"`
	U8            uint8       `json:"U8:u"`
	U16           uint16      `json:"U16:u"`
	U32           uint32      `json:"U32:u"`
	F32           float32     `json:"F32:f"`
	F64           float64     `json:"F64:f"`
	Ai8           []int8      `json:"Ai8:ai"`
	Ai16          []int16     `json:"Ai16:ai"`
	Ai32          []int32     `json:"Ai32:ai"`
	Ai64          []int64     `json:"Ai64:ai"`
	Au16          []uint16    `json:"Au16:au"`
	Au32          []uint32    `json:"Au32:au"`
	Af32          []float32   `json:"Af32:af"`
	Af64          []float64   `json:"Af64:af"`
	AAi           [][]int     `json:"AAi:aai"`
	AAf           [][]float32 `json:"AAf:aaf"`
	AAs           [][]string  `json:"AAs:aas"`
	// O             map[string]interface{}
	B bool `json:"B:d"`
}

// Foo -
type Foo struct {
	// RawObject object.Object   `json:"@"`
	Bar       string          `json:"bar:s"`
	Bars      []string        `json:"bars:as"`
	InnerFoo  *InnerFoo       `json:"inner_foo:o"`
	InnerFoos []*InnerFoo     `json:"inner_foos:ao"`
	Object    object.Object   `json:"object:o"`
	Objects   []object.Object `json:"objects:ao"`
}

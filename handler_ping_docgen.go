// Code generated by nimona.io. DO NOT EDIT.

package nimona

import (
	"github.com/vikyd/zero"
)

var _ = zero.IsZeroVal

func (t *Ping) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "test/ping"
	}

	// # t.Nonce
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["nonce"] = t.Nonce
	}

	return m
}

func (t *Ping) FromDocumentMap(m DocumentMap) {
	*t = Ping{}

	// # t.Nonce
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["nonce"].(string); ok {
			t.Nonce = v
		}
	}

}
func (t *Pong) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "test/pong"
	}

	// # t.Nonce
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["nonce"] = t.Nonce
	}

	return m
}

func (t *Pong) FromDocumentMap(m DocumentMap) {
	*t = Pong{}

	// # t.Nonce
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["nonce"].(string); ok {
			t.Nonce = v
		}
	}

}

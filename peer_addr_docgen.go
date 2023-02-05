// Code generated by nimona.io. DO NOT EDIT.

package nimona

import (
	"github.com/vikyd/zero"
)

var _ = zero.IsZeroVal

func (t *PeerAddr) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "core/node.address"
	}

	// # t.Address
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Address) {
			m["address"] = t.Address
		}
	}

	// # t.Network
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Network) {
			m["network"] = t.Network
		}
	}

	// # t.PublicKey
	//
	// Type: nimona.PublicKey, Kind: slice
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: uint8, ElemKind: uint8
	// IsElemSlice: false, IsElemStruct: false, IsElemPointer: false
	{
		if !zero.IsZeroVal(t.PublicKey) {
			m["publicKey"] = []byte(t.PublicKey)
		}
	}

	return m
}

func (t *PeerAddr) FromDocumentMap(m DocumentMap) {
	*t = PeerAddr{}

	// # t.Address
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["address"].(string); ok {
			t.Address = v
		}
	}

	// # t.Network
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["network"].(string); ok {
			t.Network = v
		}
	}

	// # t.PublicKey
	//
	// Type: nimona.PublicKey, Kind: slice
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: uint8, ElemKind: uint8
	// IsElemSlice: false, IsElemStruct: false, IsElemPointer: false
	{
		if v, ok := m["publicKey"].([]byte); ok {
			t.PublicKey = v
		}
	}

}

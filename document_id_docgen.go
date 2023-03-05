// Code generated by nimona.io. DO NOT EDIT.

package nimona

import (
	"github.com/vikyd/zero"

	"nimona.io/internal/tilde"
)

var _ = zero.IsZeroVal
var _ = tilde.NewScanner

func (t *DocumentID) Document() *Document {
	return NewDocument(t.Map())
}

func (t *DocumentID) Map() tilde.Map {
	m := tilde.Map{}

	// # t.$type
	//
	// Type: string, Kind: string, TildeKind: InvalidValueKind0
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m.Set("$type", tilde.String("core/document/id"))
	}

	// # t.DocumentHash
	//
	// Type: nimona.DocumentHash, Kind: array, TildeKind: InvalidValueKind5
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.DocumentHash) {
			m.Set("hash", tilde.Ref(t.DocumentHash[:]))
		}
	}

	return m
}

func (t *DocumentID) FromDocument(d *Document) error {
	return t.FromMap(d.Map())
}

func (t *DocumentID) FromMap(d tilde.Map) error {
	*t = DocumentID{}

	// # t.DocumentHash
	//
	// Type: nimona.DocumentHash, Kind: array, TildeKind: InvalidValueKind5
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("hash"); err == nil {
			if v, ok := v.(tilde.Ref); ok {
				copy(t.DocumentHash[:], v)
			}
		}
	}

	return nil
}

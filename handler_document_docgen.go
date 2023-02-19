// Code generated by nimona.io. DO NOT EDIT.

package nimona

import (
	"github.com/vikyd/zero"

	"nimona.io/internal/tilde"
)

var _ = zero.IsZeroVal
var _ = tilde.NewScanner

func (t *DocumentRequest) DocumentMap() *DocumentMap {
	m := tilde.Map{}

	// # t.$type
	//
	// Type: string, Kind: string, TildeKind: InvalidValueKind0
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m.Set("$type", tilde.String("core/document.request"))
	}

	// # t.DocumentID
	//
	// Type: nimona.DocumentID, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		m.Set("documentID", t.DocumentID.DocumentMap().m)
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.Metadata) {
			m.Set("$metadata", t.Metadata.DocumentMap().m)
		}
	}

	return NewDocumentMap(m)
}

func (t *DocumentRequest) FromDocumentMap(d *DocumentMap) error {
	*t = DocumentRequest{}

	// # t.DocumentID
	//
	// Type: nimona.DocumentID, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, err := d.m.Get("documentID"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := DocumentID{}
				d := NewDocumentMap(v)
				e.FromDocumentMap(d)
				t.DocumentID = e
			}
		}
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, err := d.m.Get("$metadata"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := Metadata{}
				d := NewDocumentMap(v)
				e.FromDocumentMap(d)
				t.Metadata = e
			}
		}
	}

	return nil
}
func (t *DocumentResponse) DocumentMap() *DocumentMap {
	m := tilde.Map{}

	// # t.$type
	//
	// Type: string, Kind: string, TildeKind: InvalidValueKind0
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m.Set("$type", tilde.String("core/document.response"))
	}

	// # t.Document
	//
	// Type: nimona.DocumentMap, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		m.Set("document", t.Document.DocumentMap().m)
	}

	// # t.Error
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Error) {
			m.Set("error", tilde.Bool(t.Error))
		}
	}

	// # t.ErrorDescription
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.ErrorDescription) {
			m.Set("errorDescription", tilde.String(t.ErrorDescription))
		}
	}

	// # t.Found
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m.Set("found", tilde.Bool(t.Found))
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.Metadata) {
			m.Set("$metadata", t.Metadata.DocumentMap().m)
		}
	}

	return NewDocumentMap(m)
}

func (t *DocumentResponse) FromDocumentMap(d *DocumentMap) error {
	*t = DocumentResponse{}

	// # t.Document
	//
	// Type: nimona.DocumentMap, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, err := d.m.Get("document"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				t.Document = *NewDocumentMap(v)
			}
		}
	}

	// # t.Error
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.m.Get("error"); err == nil {
			if v, ok := v.(tilde.Bool); ok {
				t.Error = bool(v)
			}
		}
	}

	// # t.ErrorDescription
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.m.Get("errorDescription"); err == nil {
			if v, ok := v.(tilde.String); ok {
				t.ErrorDescription = string(v)
			}
		}
	}

	// # t.Found
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.m.Get("found"); err == nil {
			if v, ok := v.(tilde.Bool); ok {
				t.Found = bool(v)
			}
		}
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, err := d.m.Get("$metadata"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := Metadata{}
				d := NewDocumentMap(v)
				e.FromDocumentMap(d)
				t.Metadata = e
			}
		}
	}

	return nil
}

// Code generated by nimona.io. DO NOT EDIT.

package nimona

import (
	"github.com/vikyd/zero"

	"nimona.io/tilde"
)

var _ = zero.IsZeroVal
var _ = tilde.NewScanner

func (t *DocumentPatch) Document() *Document {
	return NewDocument(t.Map())
}

func (t *DocumentPatch) Map() tilde.Map {
	m := tilde.Map{}

	// # t.$type
	//
	// Type: string, Kind: string, TildeKind: InvalidValueKind0
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m.Set("$type", tilde.String("core/stream/patch"))
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.Metadata) {
			m.Set("$metadata", t.Metadata.Map())
		}
	}

	// # t.Operations
	//
	// Type: []nimona.DocumentPatchOperation, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: nimona.DocumentPatchOperation, ElemKind: struct
	// IsElemSlice: false, IsElemStruct: true, IsElemPointer: false
	{
		if !zero.IsZeroVal(t.Operations) {
			sm := tilde.List{}
			for _, v := range t.Operations {
				if !zero.IsZeroVal(t.Operations) {
					sm = append(sm, v.Map())
				}
			}
			m.Set("operations", sm)
		}
	}

	return m
}

func (t *DocumentPatch) FromDocument(d *Document) error {
	return t.FromMap(d.Map())
}

func (t *DocumentPatch) FromMap(d tilde.Map) error {
	*t = DocumentPatch{}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, err := d.Get("$metadata"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := Metadata{}
				d := NewDocument(v)
				e.FromDocument(d)
				t.Metadata = e
			}
		}
	}

	// # t.Operations
	//
	// Type: []nimona.DocumentPatchOperation, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: nimona.DocumentPatchOperation, ElemKind: struct, ElemTildeKind: Map
	// IsElemSlice: false, IsElemStruct: true, IsElemPointer: false
	{
		sm := []DocumentPatchOperation{}
		if vs, err := d.Get("operations"); err == nil {
			if vs, ok := vs.(tilde.List); ok {
				for _, vi := range vs {
					if v, ok := vi.(tilde.Map); ok {
						e := DocumentPatchOperation{}
						d := NewDocument(v)
						e.FromDocument(d)
						sm = append(sm, e)
					}
				}
			}
		}
		if len(sm) > 0 {
			t.Operations = sm
		}
	}

	return nil
}
func (t *DocumentPatchOperation) Document() *Document {
	return NewDocument(t.Map())
}

func (t *DocumentPatchOperation) Map() tilde.Map {
	m := tilde.Map{}

	// # t.Key
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Key) {
			m.Set("key", tilde.String(t.Key))
		}
	}

	// # t.Op
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m.Set("op", tilde.String(t.Op))
	}

	// # t.Partition
	//
	// Type: []string, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: string, ElemKind: string
	// IsElemSlice: false, IsElemStruct: false, IsElemPointer: false
	{
		if !zero.IsZeroVal(t.Partition) {
			s := make(tilde.List, len(t.Partition))
			for i, v := range t.Partition {
				s[i] = tilde.String(v)
			}
			m.Set("partition", s)
		}
	}

	// # t.Path
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m.Set("path", tilde.String(t.Path))
	}

	// # t.Value
	//
	// Type: tilde.Value, Kind: interface, TildeKind: Value
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Value) {
			m.Set("value", tilde.Value(t.Value))
		}
	}

	return m
}

func (t *DocumentPatchOperation) FromDocument(d *Document) error {
	return t.FromMap(d.Map())
}

func (t *DocumentPatchOperation) FromMap(d tilde.Map) error {
	*t = DocumentPatchOperation{}

	// # t.Key
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("key"); err == nil {
			if v, ok := v.(tilde.String); ok {
				t.Key = string(v)
			}
		}
	}

	// # t.Op
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("op"); err == nil {
			if v, ok := v.(tilde.String); ok {
				t.Op = string(v)
			}
		}
	}

	// # t.Partition
	//
	// Type: []string, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: string, ElemKind: string, ElemTildeKind: String
	// IsElemSlice: false, IsElemStruct: false, IsElemPointer: false
	{
		if v, err := d.Get("partition"); err == nil {
			if v, ok := v.(tilde.List); ok {
				s := make([]string, len(v))
				for i, vi := range v {
					if vi, ok := vi.(tilde.String); ok {
						s[i] = string(vi)
					}
				}
				t.Partition = s
			}
		}
	}

	// # t.Path
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("path"); err == nil {
			if v, ok := v.(tilde.String); ok {
				t.Path = string(v)
			}
		}
	}

	// # t.Value
	//
	// Type: tilde.Value, Kind: interface, TildeKind: Value
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("value"); err == nil {
			if v, ok := v.(tilde.Value); ok {
				t.Value = tilde.Value(v)
			}
		}
	}

	return nil
}
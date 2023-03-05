// Code generated by nimona.io. DO NOT EDIT.

package nimona

import (
	"github.com/vikyd/zero"

	"nimona.io/internal/tilde"
)

var _ = zero.IsZeroVal
var _ = tilde.NewScanner

func (t *Profile) Document() *Document {
	return NewDocument(t.Map())
}

func (t *Profile) Map() tilde.Map {
	m := tilde.Map{}

	// # t.$type
	//
	// Type: string, Kind: string, TildeKind: InvalidValueKind0
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m.Set("$type", tilde.String("core/identity/profile"))
	}

	// # t.DisplayName
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.DisplayName) {
			m.Set("displayName", tilde.String(t.DisplayName))
		}
	}

	// # t.Identity
	//
	// Type: nimona.Identity, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if !zero.IsZeroVal(t.Identity) {
			m.Set("identity", t.Identity.Map())
		}
	}

	// # t.IdentityAlias
	//
	// Type: nimona.IdentityAlias, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if !zero.IsZeroVal(t.IdentityAlias) {
			m.Set("identityAlias", t.IdentityAlias.Map())
		}
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

	// # t.Repositories
	//
	// Type: []nimona.ProfileRepository, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: nimona.ProfileRepository, ElemKind: struct
	// IsElemSlice: false, IsElemStruct: true, IsElemPointer: false
	{
		if !zero.IsZeroVal(t.Repositories) {
			sm := tilde.List{}
			for _, v := range t.Repositories {
				if !zero.IsZeroVal(t.Repositories) {
					sm = append(sm, v.Map())
				}
			}
			m.Set("repositories", sm)
		}
	}

	return m
}

func (t *Profile) FromDocument(d *Document) error {
	return t.FromMap(d.Map())
}

func (t *Profile) FromMap(d tilde.Map) error {
	*t = Profile{}

	// # t.DisplayName
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("displayName"); err == nil {
			if v, ok := v.(tilde.String); ok {
				t.DisplayName = string(v)
			}
		}
	}

	// # t.Identity
	//
	// Type: nimona.Identity, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if v, err := d.Get("identity"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := Identity{}
				d := NewDocument(v)
				e.FromDocument(d)
				t.Identity = &e
			}
		}
	}

	// # t.IdentityAlias
	//
	// Type: nimona.IdentityAlias, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if v, err := d.Get("identityAlias"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := IdentityAlias{}
				d := NewDocument(v)
				e.FromDocument(d)
				t.IdentityAlias = &e
			}
		}
	}

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

	// # t.Repositories
	//
	// Type: []nimona.ProfileRepository, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: nimona.ProfileRepository, ElemKind: struct, ElemTildeKind: Map
	// IsElemSlice: false, IsElemStruct: true, IsElemPointer: false
	{
		sm := []ProfileRepository{}
		if vs, err := d.Get("repositories"); err == nil {
			if vs, ok := vs.(tilde.List); ok {
				for _, vi := range vs {
					if v, ok := vi.(tilde.Map); ok {
						e := ProfileRepository{}
						d := NewDocument(v)
						e.FromDocument(d)
						sm = append(sm, e)
					}
				}
			}
		}
		if len(sm) > 0 {
			t.Repositories = sm
		}
	}

	return nil
}
func (t *ProfileRepository) Document() *Document {
	return NewDocument(t.Map())
}

func (t *ProfileRepository) Map() tilde.Map {
	m := tilde.Map{}

	// # t.Alias
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Alias) {
			m.Set("alias", tilde.String(t.Alias))
		}
	}

	// # t.DocumentTypes
	//
	// Type: []string, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: string, ElemKind: string
	// IsElemSlice: false, IsElemStruct: false, IsElemPointer: false
	{
		s := make(tilde.List, len(t.DocumentTypes))
		for i, v := range t.DocumentTypes {
			s[i] = tilde.String(v)
		}
		m.Set("documentTypes", s)
	}

	// # t.Handle
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Handle) {
			m.Set("handle", tilde.String(t.Handle))
		}
	}

	// # t.Identity
	//
	// Type: nimona.Identity, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if !zero.IsZeroVal(t.Identity) {
			m.Set("identity", t.Identity.Map())
		}
	}

	return m
}

func (t *ProfileRepository) FromDocument(d *Document) error {
	return t.FromMap(d.Map())
}

func (t *ProfileRepository) FromMap(d tilde.Map) error {
	*t = ProfileRepository{}

	// # t.Alias
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("alias"); err == nil {
			if v, ok := v.(tilde.String); ok {
				t.Alias = string(v)
			}
		}
	}

	// # t.DocumentTypes
	//
	// Type: []string, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: string, ElemKind: string, ElemTildeKind: String
	// IsElemSlice: false, IsElemStruct: false, IsElemPointer: false
	{
		if v, err := d.Get("documentTypes"); err == nil {
			if v, ok := v.(tilde.List); ok {
				s := make([]string, len(v))
				for i, vi := range v {
					if vi, ok := vi.(tilde.String); ok {
						s[i] = string(vi)
					}
				}
				t.DocumentTypes = s
			}
		}
	}

	// # t.Handle
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("handle"); err == nil {
			if v, ok := v.(tilde.String); ok {
				t.Handle = string(v)
			}
		}
	}

	// # t.Identity
	//
	// Type: nimona.Identity, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if v, err := d.Get("identity"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := Identity{}
				d := NewDocument(v)
				e.FromDocument(d)
				t.Identity = &e
			}
		}
	}

	return nil
}

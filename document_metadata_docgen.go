// Code generated by nimona.io. DO NOT EDIT.

package nimona

import (
	"github.com/vikyd/zero"

	"nimona.io/tilde"
)

var _ = zero.IsZeroVal
var _ = tilde.NewScanner

func (t *Metadata) Document() *Document {
	return NewDocument(t.Map())
}

func (t *Metadata) Map() tilde.Map {
	m := tilde.Map{}

	// # t.Owner
	//
	// Type: nimona.Identity, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if !zero.IsZeroVal(t.Owner) {
			m.Set("owner", t.Owner.Map())
		}
	}

	// # t.Parents
	//
	// Type: []nimona.DocumentID, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: nimona.DocumentID, ElemKind: struct
	// IsElemSlice: false, IsElemStruct: true, IsElemPointer: false
	{
		if !zero.IsZeroVal(t.Parents) {
			sm := tilde.List{}
			for i, _ := range t.Parents {
				v := t.Parents[i]
				if !zero.IsZeroVal(v) {
					sm = append(sm, v.Map())
				}
			}
			m.Set("parents", sm)
		}
	}

	// # t.Permissions
	//
	// Type: []nimona.Permissions, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: nimona.Permissions, ElemKind: struct
	// IsElemSlice: false, IsElemStruct: true, IsElemPointer: false
	{
		if !zero.IsZeroVal(t.Permissions) {
			sm := tilde.List{}
			for i, _ := range t.Permissions {
				v := t.Permissions[i]
				if !zero.IsZeroVal(v) {
					sm = append(sm, v.Map())
				}
			}
			m.Set("permissions", sm)
		}
	}

	// # t.Root
	//
	// Type: nimona.DocumentID, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if !zero.IsZeroVal(t.Root) {
			m.Set("root", t.Root.Map())
		}
	}

	// # t.Sequence
	//
	// Type: uint64, Kind: uint64, TildeKind: Uint64
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Sequence) {
			m.Set("sequence", tilde.Uint64(t.Sequence))
		}
	}

	// # t.Signature
	//
	// Type: nimona.Signature, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if !zero.IsZeroVal(t.Signature) {
			m.Set("_signature", t.Signature.Map())
		}
	}

	// # t.Timestamp
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Timestamp) {
			m.Set("timestamp", tilde.String(t.Timestamp))
		}
	}

	return m
}

func (t *Metadata) FromDocument(d *Document) error {
	return t.FromMap(d.Map())
}

func (t *Metadata) FromMap(d tilde.Map) error {
	*t = Metadata{}

	// # t.Owner
	//
	// Type: nimona.Identity, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if v, err := d.Get("owner"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := Identity{}
				d := NewDocument(v)
				e.FromDocument(d)
				t.Owner = &e
			}
		}
	}

	// # t.Parents
	//
	// Type: []nimona.DocumentID, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: nimona.DocumentID, ElemKind: struct, ElemTildeKind: Map
	// IsElemSlice: false, IsElemStruct: true, IsElemPointer: false
	{
		sm := []DocumentID{}
		if vs, err := d.Get("parents"); err == nil {
			if vs, ok := vs.(tilde.List); ok {
				for _, vi := range vs {
					if v, ok := vi.(tilde.Map); ok {
						e := DocumentID{}
						d := NewDocument(v)
						e.FromDocument(d)
						sm = append(sm, e)
					}
				}
			}
		}
		if len(sm) > 0 {
			t.Parents = sm
		}
	}

	// # t.Permissions
	//
	// Type: []nimona.Permissions, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: nimona.Permissions, ElemKind: struct, ElemTildeKind: Map
	// IsElemSlice: false, IsElemStruct: true, IsElemPointer: false
	{
		sm := []Permissions{}
		if vs, err := d.Get("permissions"); err == nil {
			if vs, ok := vs.(tilde.List); ok {
				for _, vi := range vs {
					if v, ok := vi.(tilde.Map); ok {
						e := Permissions{}
						d := NewDocument(v)
						e.FromDocument(d)
						sm = append(sm, e)
					}
				}
			}
		}
		if len(sm) > 0 {
			t.Permissions = sm
		}
	}

	// # t.Root
	//
	// Type: nimona.DocumentID, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if v, err := d.Get("root"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := DocumentID{}
				d := NewDocument(v)
				e.FromDocument(d)
				t.Root = &e
			}
		}
	}

	// # t.Sequence
	//
	// Type: uint64, Kind: uint64, TildeKind: Uint64
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("sequence"); err == nil {
			if v, ok := v.(tilde.Uint64); ok {
				t.Sequence = uint64(v)
			}
		}
	}

	// # t.Signature
	//
	// Type: nimona.Signature, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: true
	{
		if v, err := d.Get("_signature"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := Signature{}
				d := NewDocument(v)
				e.FromDocument(d)
				t.Signature = &e
			}
		}
	}

	// # t.Timestamp
	//
	// Type: string, Kind: string, TildeKind: String
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("timestamp"); err == nil {
			if v, ok := v.(tilde.String); ok {
				t.Timestamp = string(v)
			}
		}
	}

	return nil
}
func (t *Permissions) Document() *Document {
	return NewDocument(t.Map())
}

func (t *Permissions) Map() tilde.Map {
	m := tilde.Map{}

	// # t.Conditions
	//
	// Type: nimona.PermissionsCondition, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		m.Set("conditions", t.Conditions.Map())
	}

	// # t.Operations
	//
	// Type: nimona.PermissionsAllow, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		m.Set("operations", t.Operations.Map())
	}

	// # t.Paths
	//
	// Type: []string, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: string, ElemKind: string
	// IsElemSlice: false, IsElemStruct: false, IsElemPointer: false
	{
		s := make(tilde.List, len(t.Paths))
		for i, v := range t.Paths {
			s[i] = tilde.String(v)
		}
		m.Set("paths", s)
	}

	return m
}

func (t *Permissions) FromDocument(d *Document) error {
	return t.FromMap(d.Map())
}

func (t *Permissions) FromMap(d tilde.Map) error {
	*t = Permissions{}

	// # t.Conditions
	//
	// Type: nimona.PermissionsCondition, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, err := d.Get("conditions"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := PermissionsCondition{}
				d := NewDocument(v)
				e.FromDocument(d)
				t.Conditions = e
			}
		}
	}

	// # t.Operations
	//
	// Type: nimona.PermissionsAllow, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, err := d.Get("operations"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := PermissionsAllow{}
				d := NewDocument(v)
				e.FromDocument(d)
				t.Operations = e
			}
		}
	}

	// # t.Paths
	//
	// Type: []string, Kind: slice, TildeKind: List
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: string, ElemKind: string, ElemTildeKind: String
	// IsElemSlice: false, IsElemStruct: false, IsElemPointer: false
	{
		if v, err := d.Get("paths"); err == nil {
			if v, ok := v.(tilde.List); ok {
				s := make([]string, len(v))
				for i, vi := range v {
					if vi, ok := vi.(tilde.String); ok {
						s[i] = string(vi)
					}
				}
				t.Paths = s
			}
		}
	}

	return nil
}
func (t *PermissionsAllow) Document() *Document {
	return NewDocument(t.Map())
}

func (t *PermissionsAllow) Map() tilde.Map {
	m := tilde.Map{}

	// # t.Add
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Add) {
			m.Set("add", tilde.Bool(t.Add))
		}
	}

	// # t.Copy
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Copy) {
			m.Set("copy", tilde.Bool(t.Copy))
		}
	}

	// # t.Move
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Move) {
			m.Set("move", tilde.Bool(t.Move))
		}
	}

	// # t.Read
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Read) {
			m.Set("read", tilde.Bool(t.Read))
		}
	}

	// # t.Remove
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Remove) {
			m.Set("remove", tilde.Bool(t.Remove))
		}
	}

	// # t.Replace
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Replace) {
			m.Set("replace", tilde.Bool(t.Replace))
		}
	}

	// # t.Test
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Test) {
			m.Set("test", tilde.Bool(t.Test))
		}
	}

	return m
}

func (t *PermissionsAllow) FromDocument(d *Document) error {
	return t.FromMap(d.Map())
}

func (t *PermissionsAllow) FromMap(d tilde.Map) error {
	*t = PermissionsAllow{}

	// # t.Add
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("add"); err == nil {
			if v, ok := v.(tilde.Bool); ok {
				t.Add = bool(v)
			}
		}
	}

	// # t.Copy
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("copy"); err == nil {
			if v, ok := v.(tilde.Bool); ok {
				t.Copy = bool(v)
			}
		}
	}

	// # t.Move
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("move"); err == nil {
			if v, ok := v.(tilde.Bool); ok {
				t.Move = bool(v)
			}
		}
	}

	// # t.Read
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("read"); err == nil {
			if v, ok := v.(tilde.Bool); ok {
				t.Read = bool(v)
			}
		}
	}

	// # t.Remove
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("remove"); err == nil {
			if v, ok := v.(tilde.Bool); ok {
				t.Remove = bool(v)
			}
		}
	}

	// # t.Replace
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("replace"); err == nil {
			if v, ok := v.(tilde.Bool); ok {
				t.Replace = bool(v)
			}
		}
	}

	// # t.Test
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("test"); err == nil {
			if v, ok := v.(tilde.Bool); ok {
				t.Test = bool(v)
			}
		}
	}

	return nil
}
func (t *PermissionsCondition) Document() *Document {
	return NewDocument(t.Map())
}

func (t *PermissionsCondition) Map() tilde.Map {
	m := tilde.Map{}

	// # t.IsOwner
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.IsOwner) {
			m.Set("isOwner", tilde.Bool(t.IsOwner))
		}
	}

	return m
}

func (t *PermissionsCondition) FromDocument(d *Document) error {
	return t.FromMap(d.Map())
}

func (t *PermissionsCondition) FromMap(d tilde.Map) error {
	*t = PermissionsCondition{}

	// # t.IsOwner
	//
	// Type: bool, Kind: bool, TildeKind: Bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, err := d.Get("isOwner"); err == nil {
			if v, ok := v.(tilde.Bool); ok {
				t.IsOwner = bool(v)
			}
		}
	}

	return nil
}
func (t *Signature) Document() *Document {
	return NewDocument(t.Map())
}

func (t *Signature) Map() tilde.Map {
	m := tilde.Map{}

	// # t.Signer
	//
	// Type: nimona.PeerKey, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		m.Set("signer", t.Signer.Map())
	}

	// # t.X
	//
	// Type: []uint8, Kind: slice, TildeKind: Bytes
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: uint8, ElemKind: uint8
	// IsElemSlice: false, IsElemStruct: false, IsElemPointer: false
	{
		m.Set("x", tilde.Bytes(t.X))
	}

	return m
}

func (t *Signature) FromDocument(d *Document) error {
	return t.FromMap(d.Map())
}

func (t *Signature) FromMap(d tilde.Map) error {
	*t = Signature{}

	// # t.Signer
	//
	// Type: nimona.PeerKey, Kind: struct, TildeKind: Map
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, err := d.Get("signer"); err == nil {
			if v, ok := v.(tilde.Map); ok {
				e := PeerKey{}
				d := NewDocument(v)
				e.FromDocument(d)
				t.Signer = e
			}
		}
	}

	// # t.X
	//
	// Type: []uint8, Kind: slice, TildeKind: Bytes
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: uint8, ElemKind: uint8, ElemTildeKind: InvalidValueKind0
	// IsElemSlice: false, IsElemStruct: false, IsElemPointer: false
	{
		if v, err := d.Get("x"); err == nil {
			if v, ok := v.(tilde.Bytes); ok {
				t.X = []byte(v)
			}
		}
	}

	return nil
}

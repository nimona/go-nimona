package object

import (
	"strings"

	"nimona.io/internal/encoding/base58"
)

type (
	// Member
	Member struct {
		Name     string
		Value    interface{}
		TypeHint TypeHint
	}

	// Object for everything f12n
	Object struct {
		Members map[string]*Member
	}
)

// HintedName returns the member's name and hint
func (m *Member) HintedName() string {
	if m.TypeHint == HintUndefined {
		return m.Name
	}
	return m.Name + ":" + string(m.TypeHint)
}

// FromBytes returns an object from a cbor byte stream
// TODO: Remove and move to a generic codec
func FromBytes(b []byte) (*Object, error) {
	m := map[string]interface{}{}
	if err := UnmarshalSimple(b, &m); err != nil {
		return nil, err
	}

	o := FromMap(m)
	return o, nil
}

// New returns an object from a map
func New() *Object {
	return &Object{
		Members: map[string]*Member{},
	}
}

// FromMap returns an object from a map
func FromMap(m map[string]interface{}) *Object {
	o := New()
	o.FromMap(m)
	return o
}

// FromMap inits the object from a map
func (o *Object) FromMap(m map[string]interface{}) error {
	// TODO figure out what should error here
	for k, v := range m {
		o.SetRaw(k, v)
	}
	return nil
}

// Hash returns the object's hash
func (o Object) Hash() []byte {
	return Hash(&o)
}

// HashBase58 returns the object's hash base58 encoded
func (o Object) HashBase58() string {
	return base58.Encode(Hash(&o))
}

// ToMap returns the object as a map
func (o Object) ToMap() map[string]interface{} {
	r := map[string]interface{}{}
	for _, m := range o.Members {
		// TODO check the type hint first maybe?
		if io, ok := m.Value.(*Object); ok {
			r[m.HintedName()] = io.ToMap()
		} else {
			r[m.HintedName()] = m.Value
		}
	}
	return r
}

// ToPlainMap returns the object as a map, without adding type hints on the keys
func (o Object) ToPlainMap() map[string]interface{} {
	r := map[string]interface{}{}
	for _, m := range o.Members {
		// TODO check the type hint first maybe?
		if io, ok := m.Value.(*Object); ok {
			r[m.Name] = io.ToMap()
		} else {
			r[m.Name] = m.Value
		}
	}
	return r
}

// GetType returns the object's type
func (o Object) GetType() string {
	if v, ok := o.GetRaw("@ctx").(string); ok {
		return v
	}
	return ""
}

// SetType sets the object's type
func (o Object) SetType(v string) {
	o.SetRaw("@ctx", v)
}

// GetSignature returns the object's signature, or nil
func (o Object) GetSignature() *Object {
	if v, ok := o.GetRaw("@signature").(*Object); ok {
		return v
	}
	return nil
}

// SetSignature sets the object's signature
func (o Object) SetSignature(v *Object) {
	o.SetRaw("@signature", v)
}

// GetAuthorityKey returns the object's creator, or nil
func (o Object) GetAuthorityKey() *Object {
	if v, ok := o.GetRaw("@authority").(*Object); ok {
		return v
	}
	return nil
}

// SetAuthorityKey sets the object's creator
func (o Object) SetAuthorityKey(v *Object) {
	o.SetRaw("@authority", v)
}

// GetMandate returns the object's mandate, or nil
func (o Object) GetMandate() *Object {
	if v, ok := o.GetRaw("@mandate").(*Object); ok {
		return v
	}
	return nil
}

// SetMandate sets the object's mandate
func (o Object) SetMandate(v *Object) {
	o.SetRaw("@mandate", v)
}

// GetSignerKey returns the object's signer, or nil
func (o Object) GetSignerKey() *Object {
	if v, ok := o.GetRaw("@signer").(*Object); ok {
		return v
	}
	return nil
}

// SetSignerKey sets the object's signer
func (o Object) SetSignerKey(v *Object) {
	o.SetRaw("@signer", v)
}

// GetPolicy returns the object's policy, or nil
func (o Object) GetPolicy() *Object {
	if v, ok := o.GetRaw("@policy").(*Object); ok {
		return v
	}
	return nil
}

// SetPolicy sets the object's policy
func (o Object) SetPolicy(v *Object) {
	o.SetRaw("@policy", v)
}

// GetParents returns the object's parent refs
func (o Object) GetParents() []string {
	if v, ok := o.GetRaw("@parents").([]string); ok {
		return v
	}
	return nil
}

// SetParents sets the object's parents
func (o Object) SetParents(v []string) {
	o.SetRaw("@parents", v)
}

// GetRaw -
func (o Object) GetRaw(lk string) interface{} {
	for _, m := range o.Members {
		if m.Name == lk {
			return m.Value
		}
	}

	return nil
}

// SetRaw -
func (o *Object) SetRaw(n string, v interface{}) {
	ct := ""
	np := strings.Split(n, ":")
	if len(np) > 1 {
		ct = np[1]
	}
	n = np[0]

	if mv, ok := v.(map[string]interface{}); ok {
		tv := FromMap(mv)
		if tv.GetType() != "" {
			v = tv
		}
	}

	var nv interface{}
	if oi, ok := v.(*Object); ok {
		nv = oi
	} else if oi, ok := v.(objectable); ok {
		nv = oi.ToObject()
	} else if m, ok := v.(map[string]interface{}); ok {
		nv = FromMap(m)
	} else {
		nv = v
	}

	t := DeduceTypeHint(v)
	if ct != "" && ct != string(t) {
		// TODO: should we error or something if the given type hint does not
		// match the actual?
	}

	m := &Member{
		Name:     n,
		Value:    nv,
		TypeHint: t,
	}

	if o.Members == nil {
		o.Members = map[string]*Member{}
	}

	o.Members[m.Name] = m
}

package encoding

import (
	"nimona.io/go/base58"
)

// Typed interface
type Typed interface {
	Type() string
}

// Object for everything f12n
type Object map[string]interface{}

// NewObjectFromBytes returns an object from a cbor byte stream
func NewObjectFromBytes(b []byte) (*Object, error) {
	m := map[string]interface{}{}
	if err := UnmarshalSimple(b, &m); err != nil {
		return nil, err
	}

	o := NewObjectFromMap(m)
	return o, nil
}

// NewObject returns an object from a map
func NewObject() *Object {
	return &Object{}
}

// NewObjectFromMap returns an object from a map
func NewObjectFromMap(m map[string]interface{}) *Object {
	o := NewObject()
	o.FromMap(m)
	return o
}

// FromMap inits the object from a map
func (o Object) FromMap(m map[string]interface{}) {
	for k, v := range m {
		o.SetRaw(k, v)
	}
}

// Hash returns the object's hash
func (o Object) Hash() []byte {
	return Hash(&o)
}

// HashBase58 returns the object's hash base58 encoded
func (o Object) HashBase58() string {
	return base58.Encode(Hash(&o))
}

// Map returns the object as a map
func (o Object) Map() map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range o {
		// TODO check the type hint first maybe?
		if o, ok := v.(*Object); ok {
			m[k] = o.Map()
		} else {
			m[k] = v
		}
	}
	return m
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

// GetRaw -
func (o Object) GetRaw(lk string) interface{} {
	// TODO(geoah) do we need to verify type if k has hint?
	lk = getCleanKeyName(lk)
	for k, v := range o {
		if getCleanKeyName(k) == lk {
			return v
		}
	}

	return nil
}

// SetRaw -
func (o Object) SetRaw(k string, v interface{}) {
	et := getFullType(k)
	if et == "" {
		k += ":" + GetHintFromType(v)
	}

	if mv, ok := v.(map[string]interface{}); ok {
		tv := NewObjectFromMap(mv)
		if tv.GetType() != "" {
			v = tv
		}
	}

	if oi, ok := v.(*Object); ok {
		o[k] = oi
	} else if oi, ok := v.(objectable); ok {
		o[k] = oi.ToObject()
	} else if m, ok := v.(map[string]interface{}); ok {
		o[k] = NewObjectFromMap(m)
	} else {
		o[k] = v
	}
}

package encoding

import (
	"fmt"

	"nimona.io/go/base58"
)

// Typed interface
type Typed interface {
	Type() string
}

// Object for everything f12n
type Object struct {
	data      map[string]interface{}
	bytes     []byte
	ctx       string
	policy    *Object
	authority *Object
	signer    *Object
	signature *Object
}

// NewObjectFromBytes returns an object from a cbor byte stream
func NewObjectFromBytes(b []byte) (*Object, error) {
	m := map[string]interface{}{}
	if err := UnmarshalSimple(b, &m); err != nil {
		return nil, err
	}

	o := NewObjectFromMap(m)
	o.bytes = b
	return o, nil
}

// NewObjectFromStruct returns an object from a struct
// func NewObjectFromStruct(v interface{}) (*Object, error) {
// 	b, err := MarshalSimple(v)
// 	if err != nil {
// 		return nil, err
// 	}

// 	o, err := NewObjectFromBytes(b)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if vt, ok := v.(Typed); ok {
// 		o.SetType(vt.Type())
// 	}

// 	return o, nil
// }

// NewObject returns an object from a map
func NewObject() *Object {
	o := &Object{
		data: map[string]interface{}{},
	}
	return o
}

// NewObjectFromMap returns an object from a map
func NewObjectFromMap(m map[string]interface{}) *Object {
	o := NewObject()
	o.FromMap(m)
	return o
}

// FromMap inits the object from a map
func (o *Object) FromMap(m map[string]interface{}) {
	for k, v := range m {
		o.SetRaw(k, v)
	}
}

// Hash returns the object's hash
func (o *Object) Hash() []byte {
	return Hash(o)
}

// HashBase58 returns the object's hash base58 encoded
func (o *Object) HashBase58() string {
	return base58.Encode(Hash(o))
}

// Map returns the object as a map
func (o *Object) Map() map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range o.data {
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
func (o *Object) GetType() string {
	return o.ctx
}

// SetType sets the object's type
func (o *Object) SetType(v string) {
	o.SetRaw("@ctx:s", v)
}

// GetSignature returns the object's signature, or nil
func (o *Object) GetSignature() *Object {
	return o.signature
}

// SetSignature sets the object's signature
func (o *Object) SetSignature(v *Object) {
	o.SetRaw("@sig:O", v)
}

// GetAuthorityKey returns the object's creator, or nil
func (o *Object) GetAuthorityKey() *Object {
	return o.authority
}

// SetAuthorityKey sets the object's creator
func (o *Object) SetAuthorityKey(v *Object) {
	o.SetRaw("@authority:O", v)
}

// GetSignerKey returns the object's signer, or nil
func (o *Object) GetSignerKey() *Object {
	return o.signer
}

// SetSignerKey sets the object's signer
func (o *Object) SetSignerKey(v *Object) {
	o.SetRaw("@signer:O", v)
}

// GetPolicy returns the object's policy, or nil
func (o *Object) GetPolicy() *Object {
	return o.policy
}

// SetPolicy sets the object's policy
func (o *Object) SetPolicy(v *Object) {
	o.SetRaw("@policy:O", v)
}

// GetRaw -
func (o *Object) GetRaw(lk string) interface{} {
	// TODO(geoah) do we need to verify type if k has hint?
	lk = getCleanKeyName(lk)
	for k, v := range o.data {
		if getCleanKeyName(k) == lk {
			return v
		}
	}

	return nil
}

// SetRaw -
func (o *Object) SetRaw(k string, v interface{}) {
	// add type hint if not already set
	et := getFullType(k)
	if et == "" {
		k += ":" + GetHintFromType(v)
	}

	if mv, ok := v.(map[string]interface{}); ok {
		if t, ok := mv["@ctx:s"]; ok && t != "" {
			v = NewObjectFromMap(mv)
		}
	}

	// add the attribute in the data map
	o.data[k] = v

	// check if this is a magic attribute and set it
	ck := getCleanKeyName(k)
	switch ck {
	case "@ctx":
		t, ok := v.(string)
		if !ok {
			panic(fmt.Errorf("invalid type %T for @ctx", v))
		}
		o.ctx = t
	case "@policy":
		if oi, ok := v.(*Object); ok {
			o.policy = oi
		} else if oi, ok := v.(objectable); ok {
			o.policy = oi.ToObject()
		} else if m, ok := v.(map[string]interface{}); ok {
			o.policy = NewObjectFromMap(m)
		} else {
			panic(fmt.Errorf("invalid type %T for @policy", v))
		}
	case "@authority":
		if oi, ok := v.(*Object); ok {
			o.authority = oi
		} else if oi, ok := v.(objectable); ok {
			o.authority = oi.ToObject()
		} else if m, ok := v.(map[string]interface{}); ok {
			o.authority = NewObjectFromMap(m)
		} else {
			panic(fmt.Errorf("invalid type %T for @authority", v))
		}
	case "@signer":
		if oi, ok := v.(*Object); ok {
			o.signer = oi
		} else if oi, ok := v.(objectable); ok {
			o.signer = oi.ToObject()
		} else if m, ok := v.(map[string]interface{}); ok {
			o.signer = NewObjectFromMap(m)
		} else {
			panic(fmt.Errorf("invalid type %T for @signer", v))
		}
	case "@sig:O":
		if oi, ok := v.(*Object); ok {
			o.signature = oi
		} else if oi, ok := v.(objectable); ok {
			o.signature = oi.ToObject()
		} else if m, ok := v.(map[string]interface{}); ok {
			o.signature = NewObjectFromMap(m)
		} else {
			panic(fmt.Errorf("invalid type %T for @sig", v))
		}
	}
}

// Unmarshal the object into a given interface
func (o *Object) Unmarshal(v interface{}) error {
	if o.bytes == nil || len(o.bytes) == 0 {
		b, err := MarshalSimple(o.data)
		if err != nil {
			return err
		}
		o.bytes = b
	}

	return UnmarshalSimple(o.bytes, v)
}

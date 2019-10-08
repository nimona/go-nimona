package object

import (
	"encoding/json"
	"fmt"

	"nimona.io/internal/encoding/base58"
)

type (
	// Hash []byte
	// Object for everything f12n
	Object map[string]interface{}
)

// New returns an object from a map
func New() Object {
	return Object{}
}

// FromMap returns an object from a map
func FromMap(m map[string]interface{}) Object {
	return Object(m)
}

// FromMap inits the object from a map
func (o *Object) FromMap(m map[string]interface{}) error {
	v := Object(m)
	*o = v
	return nil
}

// ToObject simply returns a copy of the object
// This is mostly a hack for generated objects
func (o Object) ToObject() Object {
	return o.Copy()
}

// Hash returns the object's hash
func (o Object) Hash() *Hash {
	h, err := NewHash(o)
	if err != nil {
		panic(err)
	}

	return h
}

// ToMap returns the object as a map
func (o Object) ToMap() map[string]interface{} {
	return map[string]interface{}(o)
}

// GetType returns the object's type
func (o Object) GetType() string {
	if v, ok := o.Get("@ctx:s").(string); ok {
		return v
	}
	return ""
}

// SetType sets the object's type
func (o Object) SetType(v string) {
	o.Set("@ctx:s", v)
}

// GetSignature returns the object's signature, or nil
func (o Object) GetSignature() *Object {
	return o.getObject("@signature:o")
}

// SetSignature sets the object's signature
func (o Object) SetSignature(v Object) {
	o.Set("@signature:o", v)
}

// GetPolicy returns the object's policy, or nil
func (o Object) GetPolicy() *Object {
	return o.getObject("@policy:o")
}

func (o Object) getObject(k string) *Object {
	v := o.Get(k)
	switch o := v.(type) {
	case Object:
		return &o
	case map[string]interface{}:
		ov := Object(o)
		return &ov
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v := range o {
			m[k.(string)] = v
		}
		ov := Object(m)
		return &ov
	}
	return nil
}

// SetPolicy sets the object's policy
func (o Object) SetPolicy(v Object) {
	o.Set("@policy:o", v)
}

// GetParents returns the object's parent refs
func (o Object) GetParents() []string {
	// TODO can we use mapstructure or something else to do this?
	if v, ok := o.Get("@parents:as").([]string); ok {
		return v
	}
	if v, ok := o.Get("@parents:as").([]interface{}); ok {
		parents := []string{}
		for _, p := range v {
			ps, ok := p.(string)
			if !ok {
				continue
			}
			parents = append(parents, ps)
		}
		return parents
	}
	return nil
}

// SetParents sets the object's parents
func (o Object) SetParents(v []string) {
	o.Set("@parents:as", v)
}

// GetRoot returns the object's root
func (o Object) GetRoot() string {
	if v, ok := o.Get("@root:s").(string); ok {
		return v
	}
	return ""
}

// Get -
func (o Object) Get(lk string) interface{} {
	return o[lk]
}

// Set -
func (o Object) Set(k string, v interface{}) {
	if ov, ok := v.(Object); ok {
		v = ov.ToMap()
	}
	map[string]interface{}(o)[k] = v
}

func (o Object) Compact() (string, error) {
	j, err := json.Marshal(o.ToMap())
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"*%s.json",
		base58.Encode(j),
	), nil
}

// Copy creates a copy of the original object
func (o Object) Copy() Object {
	return Copy(o)
}

// Copy object
func Copy(f Object) Object {
	t := Object{}
	err := t.FromMap(f.ToMap())
	if err != nil {
		panic(err)
	}
	return t
}

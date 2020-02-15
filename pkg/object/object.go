package object

import (
	"strings"

	"github.com/mitchellh/mapstructure"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/immutable"
)

type (
	// Object for everything f12n
	Object struct {
		Header Header
		Data   immutable.Map
	}
	// Header for object
	Header struct {
		Type      string             `json:"type:s,omitempty" mapstructure:"type:s,omitempty"`
		Stream    Hash               `json:"stream:s,omitempty" mapstructure:"stream:s,omitempty"`
		Parents   []Hash             `json:"parents:as,omitempty" mapstructure:"parents:as,omitempty"`
		Policy    Policy             `json:"policy:o,omitempty" mapstructure:"policy:o,omitempty"`
		Signature Signature          `json:"_signature:o,omitempty" mapstructure:"_signature:o,omitempty"`
		Owners    []crypto.PublicKey `json:"owners:as,omitempty" mapstructure:"owners:as,omitempty"`
	}
	// Policy for object
	Policy struct {
		Subjects  []string `json:"subjects:as,omitempty" mapstructure:"subjects:as,omitempty"`
		Resources []string `json:"resources:as,omitempty" mapstructure:"resources:as,omitempty"`
		Actions   []string `json:"actions:as,omitempty" mapstructure:"actions:as,omitempty"`
		Effect    string   `json:"effect:s,omitempty" mapstructure:"effect:s,omitempty"`
	}
)

func (v *Object) SetType(t string) {
	v.Header.Type = t
}

func (v Object) GetType() string {
	return v.Header.Type
}

func (v Signature) IsEmpty() bool {
	return v.Signer.IsEmpty()
}

func (v Policy) IsEmpty() bool {
	return len(v.Subjects) == 0
}

// FromMap returns an object from a map
func FromMap(m map[string]interface{}) Object {
	n, err := normalizeObject(m)
	if err != nil {
		panic(err)
	}

	h := Header{}
	if mh, ok := n["header:o"]; ok {
		mapstructure.Decode(mh, &h) // nolint: errcheck
	}

	o := Object{
		Header: h,
		Data:   immutable.Map{},
	}

	if md, ok := n["data:o"]; ok {
		o.Data = immutable.AnyToValue(":o", md).(immutable.Map)
	}

	return o
}

// ToObject returns the same object, this is a helper method for codegen
func (o Object) ToObject() Object {
	return o
}

// IsEmpty returns whether the object is empty
func (o Object) IsEmpty() bool {
	return o.Data.PrimitiveHinted() == nil && o.Header.Type == ""
}

// ToMap returns the object as a map
func (o Object) ToMap() map[string]interface{} {
	h := map[string]interface{}{}
	h["type:s"] = o.Header.Type
	if !o.Header.Stream.IsEmpty() {
		h["stream:s"] = o.Header.Stream.String()
	}
	if len(o.Header.Parents) != 0 {
		ps := make([]string, len(o.Header.Parents))
		for i, p := range o.Header.Parents {
			ps[i] = p.String()
		}
		h["parents:as"] = ps
	}
	if !o.Header.Policy.IsEmpty() {
		p := map[string]interface{}{}
		if len(o.Header.Policy.Subjects) > 0 {
			p["subjects:as"] = o.Header.Policy.Subjects
		}
		if len(o.Header.Policy.Resources) > 0 {
			p["resources:as"] = o.Header.Policy.Resources
		}
		if len(o.Header.Policy.Actions) > 0 {
			p["actions:as"] = o.Header.Policy.Actions
		}
		if len(o.Header.Policy.Effect) > 0 {
			p["effect:s"] = o.Header.Policy.Effect
		}
		h["policy:o"] = p
	}
	if !o.Header.Signature.IsEmpty() {
		h["_signature:o"] = o.Header.Signature.ToMap()
	}
	if len(o.Header.Owners) != 0 {
		os := make([]string, len(o.Header.Owners))
		for i, owner := range o.Header.Owners {
			os[i] = owner.String()
		}
		h["owners:as"] = os
	}
	n, err := normalizeObject(map[string]interface{}{
		"header:o": h,
	})
	p := o.Data.PrimitiveHinted()
	if p != nil {
		n["data:o"] = p.(map[string]interface{})
	}
	if err != nil {
		panic(err)
	}
	return n
}

// Get -
func (o Object) Get(k string) interface{} {
	// remove hint from key
	ps := strings.Split(k, ":")
	if len(ps) > 1 {
		k = ps[0]
	}
	v := o.Data.Value(k)
	if v == nil {
		return nil
	}
	return v.PrimitiveHinted()
}

// Set -
func (o *Object) Set(k string, v interface{}) {
	o.Data = o.Data.Set(k, immutable.AnyToValue(k, v))
}

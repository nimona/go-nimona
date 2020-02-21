package object

import (
	"github.com/mitchellh/mapstructure"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/immutable"
)

type (
	// Object for everything f12n
	Object immutable.Map
	// Policy for object
	Policy struct {
		Subjects  []string `json:"subjects:as,omitempty" mapstructure:"subjects:as,omitempty"`
		Resources []string `json:"resources:as,omitempty" mapstructure:"resources:as,omitempty"`
		Actions   []string `json:"actions:as,omitempty" mapstructure:"actions:as,omitempty"`
		Effect    string   `json:"effect:s,omitempty" mapstructure:"effect:s,omitempty"`
	}
)

func (v Signature) IsEmpty() bool {
	return v.Signer.IsEmpty()
}

func (v Policy) IsEmpty() bool {
	return len(v.Subjects) == 0
}

func (o Object) set(k string, v immutable.Value) Object {
	return Object(immutable.Map(o).Set(k, v))
}

func (o Object) data() immutable.Map {
	data := immutable.Map(o).Value("data:o")
	if data == nil {
		return immutable.Map{}
	}

	mdata, ok := data.(immutable.Map)
	if !ok {
		return immutable.Map{}
	}

	return mdata
}

// TODO(geoah) don't use primitives for header values

func (o Object) GetType() string {
	im := immutable.Map(o).Value("type:s")
	if im == nil {
		return ""
	}
	v, ok := im.PrimitiveHinted().(string)
	if !ok {
		return ""
	}
	return v
}

func (o Object) SetType(v string) Object {
	return o.set("type:s", immutable.String(v))
}

func (o Object) SetStream(v Hash) Object {
	return o.set("stream:s", immutable.String(v.String()))
}

func (o Object) GetStream() Hash {
	im := immutable.Map(o).Value("stream:s")
	if im == nil {
		return ""
	}
	v, ok := im.PrimitiveHinted().(string)
	if !ok {
		return ""
	}
	return Hash(v)
}

func (o Object) SetParents(hashes []Hash) Object {
	v := immutable.List{}
	for _, hash := range hashes {
		v = v.Append(immutable.String(hash.String()))
	}
	return o.set("parents:as", v)
}

func (o Object) GetParents() []Hash {
	im := immutable.Map(o).Value("parents:as")
	if im == nil {
		return []Hash{}
	}
	v, ok := im.PrimitiveHinted().([]string)
	if !ok {
		return []Hash{}
	}
	ps := make([]Hash, len(v))
	for i, p := range v {
		ps[i] = Hash(p)
	}
	return ps
}

func (o Object) SetPolicy(policy Policy) Object {
	v := immutable.Map{}
	if len(policy.Subjects) > 0 {
		v = v.Set("subjects:as", immutable.AnyToValue(":as", policy.Subjects))
	}
	if len(policy.Resources) > 0 {
		v = v.Set("resources:as", immutable.AnyToValue(":as", policy.Resources))
	}
	if len(policy.Actions) > 0 {
		v = v.Set("actions:as", immutable.AnyToValue(":as", policy.Actions))
	}
	if len(policy.Effect) > 0 {
		v = v.Set("effect:s", immutable.AnyToValue(":s", policy.Effect))
	}
	if v.IsEmpty() {
		return o
	}
	return o.set("policy:o", v)
}

func (o Object) GetPolicy() Policy {
	im := immutable.Map(o).Value("policy:o")
	if im == nil {
		return Policy{}
	}
	v, ok := im.PrimitiveHinted().(map[string]interface{})
	if !ok {
		return Policy{}
	}
	p := Policy{}
	mapstructure.Decode(v, &p) // nolint: errcheck
	return p
}

func (o Object) SetSignature(v Signature) Object {
	return o.set("_signature:o", immutable.AnyToValue(":o", v.ToMap()))
}

func (o Object) GetSignature() Signature {
	im := immutable.Map(o).Value("_signature:o")
	if im == nil {
		return Signature{}
	}
	v, ok := im.PrimitiveHinted().(map[string]interface{})
	if !ok {
		return Signature{}
	}
	s := Signature{}
	mapstructure.Decode(v, &s) // nolint: errcheck
	return s
}

func (o Object) SetOwners(owners []crypto.PublicKey) Object {
	v := immutable.List{}
	for _, owner := range owners {
		v = v.Append(immutable.String(owner.String()))
	}
	return o.set("owners:as", v)
}

func (o Object) GetOwners() []crypto.PublicKey {
	im := immutable.Map(o).Value("owners:as")
	if im == nil {
		return []crypto.PublicKey{}
	}
	v, ok := im.PrimitiveHinted().([]string)
	if !ok {
		return []crypto.PublicKey{}
	}
	os := make([]crypto.PublicKey, len(v))
	for i, p := range v {
		os[i] = crypto.PublicKey(p)
	}
	return os
}

// FromMap returns an object from a map
func FromMap(m map[string]interface{}) Object {
	if len(m) == 0 {
		return Object{}
	}

	n, err := normalizeObject(m)
	if err != nil {
		panic(err)
	}

	return Object(
		immutable.AnyToValue(":o", n).(immutable.Map),
	)
}

// ToObject returns the same object, this is a helper method for codegen
func (o Object) ToObject() Object {
	return o
}

// IsEmpty returns whether the object is empty
func (o Object) IsEmpty() bool {
	return immutable.Map(o).IsEmpty()
}

// ToMap returns the object as a map
func (o Object) ToMap() map[string]interface{} {
	im := immutable.Map(o)
	if im.IsEmpty() {
		return map[string]interface{}{}
	}
	return im.PrimitiveHinted().(map[string]interface{})
}

// Get -
func (o Object) Get(k string) interface{} {
	// remove hint from key
	// ps := strings.Split(k, ":")
	// if len(ps) > 1 {
	// 	k = ps[0]
	// }
	v := o.data().Value(k)
	if v == nil {
		return nil
	}
	return v.PrimitiveHinted()
}

// Set -
func (o Object) Set(k string, v interface{}) Object {
	data := o.data()
	if iv, ok := v.(immutable.Value); ok {
		data = data.Set(k, iv)
	} else {
		data = data.Set(k, immutable.AnyToValue(k, v))
	}
	return Object(
		immutable.Map(o).Set("data:o", data),
	)
}

func (o Object) Raw() immutable.Map {
	return immutable.Map(o)
}

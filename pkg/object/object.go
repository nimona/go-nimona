package object

import (
	"github.com/mitchellh/mapstructure"

	"nimona.io/pkg/crypto"
)

type (
	// Object for everything f12n
	Object Map
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

func (o Object) data() Map {
	data := Map(o).Value("content:m")
	if data == nil {
		return Map{}
	}

	mdata, ok := data.(Map)
	if !ok {
		return Map{}
	}

	return mdata
}

// TODO(geoah) don't use primitives for header values

func (o Object) GetType() string {
	im := Map(o).Value("type:s")
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
	return o.set("type:s", String(v))
}

func (o Object) SetStream(v Hash) Object {
	return o.set("stream:s", String(v.String()))
}

func (o Object) GetStream() Hash {
	im := Map(o).Value("stream:s")
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
	v := List{}
	for _, hash := range hashes {
		v = v.Append(String(hash.String()))
	}
	return o.set("parents:as", v)
}

func (o Object) GetParents() []Hash {
	im := Map(o).Value("parents:as")
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
	v := Map{}
	if len(policy.Subjects) > 0 {
		v = v.Set("subjects:as", AnyToValue(":as", policy.Subjects))
	}
	if len(policy.Resources) > 0 {
		v = v.Set("resources:as", AnyToValue(":as", policy.Resources))
	}
	if len(policy.Actions) > 0 {
		v = v.Set("actions:as", AnyToValue(":as", policy.Actions))
	}
	if len(policy.Effect) > 0 {
		v = v.Set("effect:s", AnyToValue(":s", policy.Effect))
	}
	if v.IsEmpty() {
		return o
	}
	return o.set("policy:m", v)
}

func (o Object) GetPolicy() Policy {
	im := Map(o).Value("policy:m")
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

func (o Object) AddSignature(vs ...Signature) Object {
	sigs := List{}
	if os := o.Get("_signatures:am"); os != nil {
		if ol, ok := os.(List); ok && ol.Length() > 0 {
			sigs = ol
		}
	}
	for _, v := range vs {
		sigs = sigs.Append(AnyToValue(":m", v.ToMap()))
	}
	return o.set("_signatures:am", sigs)
}

func immutableMapToSignature(im Map) Signature {
	if im.IsEmpty() {
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

func (o Object) GetSignatures() []Signature {
	sigs := []Signature{}
	if os := o.get("_signatures:am"); os != nil {
		if ol, ok := os.(List); ok && ol.Length() > 0 {
			ol.Iterate(func(v Value) bool {
				m, ok := v.(Map)
				if !ok {
					return true
				}
				sigs = append(sigs, immutableMapToSignature(m))
				return true
			})
		}
	}
	return sigs
}

func (o Object) SetOwners(owners []crypto.PublicKey) Object {
	v := List{}
	for _, owner := range owners {
		v = v.Append(String(owner.String()))
	}
	return o.set("owners:as", v)
}

func (o Object) GetOwners() []crypto.PublicKey {
	im := Map(o).Value("owners:as")
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
		AnyToValue(":m", n).(Map),
	)
}

// ToObject returns the same object, this is a helper method for codegen
func (o Object) ToObject() Object {
	return o
}

// IsEmpty returns whether the object is empty
func (o Object) IsEmpty() bool {
	return Map(o).IsEmpty()
}

// ToMap returns the object as a map
func (o Object) ToMap() map[string]interface{} {
	im := Map(o)
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
	if iv, ok := v.(Value); ok {
		data = data.Set(k, iv)
	} else {
		data = data.Set(k, AnyToValue(k, v))
	}
	return Object(
		Map(o).Set("content:m", data),
	)
}

func (o Object) Hash() Hash {
	return Map(o).Hash()
}

func (o Object) Raw() Map {
	return Map(o)
}

func (o Object) set(k string, v Value) Object {
	return Object(Map(o).Set(k, v))
}

func (o Object) get(k string) Value {
	return Map(o).Value(k)
}

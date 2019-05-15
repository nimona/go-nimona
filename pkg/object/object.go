package object

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/structs"

	"nimona.io/internal/encoding/base58"
)

func init() {
	structs.DefaultTagName = "json"
}

type (
	// Member -
	Member struct {
		Name     string
		Value    interface{}
		TypeHint string
	}

	// Object for everything f12n
	Object struct {
		Members map[string]*Member
	}
)

// HintedName returns the member's name and hint
func (m *Member) HintedName() string {
	if m.TypeHint == HintUndefined.String() {
		return m.Name
	}
	return m.Name + ":" + m.TypeHint
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

// ToObject simply returns a copy of the object
// This is mostly a hack for generated objects
func (o *Object) ToObject() *Object {
	return o.Copy()
}

// Hash returns the object's hash
func (o Object) Hash() []byte {
	return Hash(&o)
}

// HashBase58 returns the object's hash base58 encoded
func (o Object) HashBase58() string {
	return base58.Encode(Hash(&o))
}

// CompactHash returns the object's hash in its (not really) compact format
func (o Object) CompactHash() string {
	return "&" + base58.Encode(Hash(&o)) + ".b.oh/v1"
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
	v, ok := o.GetRaw("@signature").(map[string]interface{})
	if !ok {
		return nil
	}
	vo := &Object{}
	err := vo.FromMap(v)
	if err != nil {
		return nil
	}
	return vo
}

// SetSignature sets the object's signature
func (o Object) SetSignature(v *Object) {
	o.SetRaw("@signature", v)
}

// GetPolicy returns the object's policy, or nil
func (o Object) GetPolicy() *Object {
	v, ok := o.GetRaw("@policy").(map[string]interface{})
	if !ok {
		return nil
	}
	vo := &Object{}
	err := vo.FromMap(v)
	if err != nil {
		return nil
	}
	return vo
}

// SetPolicy sets the object's policy
func (o Object) SetPolicy(v *Object) {
	o.SetRaw("@policy", v)
}

// GetParents returns the object's parent refs
func (o Object) GetParents() []string {
	// TODO can we use mapstructure or something else to do this?
	if v, ok := o.GetRaw("@parents").([]string); ok {
		return v
	}
	if v, ok := o.GetRaw("@parents").([]interface{}); ok {
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

func raw(v interface{}) (nv interface{}) {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice:
		rv := reflect.ValueOf(v)
		// TODO this captures more than it really should, not sure if we even
		// need it any more
		ns := []interface{}{}
		for i := 0; i < rv.Len(); i++ {
			ns = append(ns, raw(rv.Index(i).Interface()))
		}
		nv = ns
	case reflect.Struct:
		if oi, ok := v.(Object); ok {
			nv = oi.ToMap()
		} else if oi, ok := v.(objectable); ok {
			nv = oi.ToObject().ToMap()
		} else if m, ok := v.(map[string]interface{}); ok {
			// TODO waste of resources, just used to make sure nested maps are typed
			nv = FromMap(m).ToMap()
		} else {
			nv = structs.Map(v)
		}
	case reflect.Ptr:
		nv = raw(reflect.ValueOf(v).Elem().Interface())
	default:
		nv = v
	}
	return nv
}

// SetRaw -
func (o *Object) SetRaw(n string, v interface{}) {
	ct := ""
	np := strings.Split(n, ":")
	if len(np) > 1 {
		ct = np[1]
	}
	n = np[0]

	nv := raw(v)

	t := DeduceTypeHint(v)
	if ct == "" {
		ct = string(t)
	} else if ct != "" && ct != string(t) {
		// TODO: should we error or something if the given type hint does not
		// match the actual?
	}

	m := &Member{
		Name:     n,
		Value:    nv,
		TypeHint: ct,
	}

	if o.Members == nil {
		o.Members = map[string]*Member{}
	}

	o.Members[m.Name] = m
}

func (o *Object) Compact() (string, error) {
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
func (o *Object) Copy() *Object {
	return Copy(o)
}

// Copy object
func Copy(f *Object) *Object {
	t := &Object{}
	err := t.FromMap(f.ToMap())
	if err != nil {
		panic(err)
	}
	return t
}

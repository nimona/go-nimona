package object

type (
	// Object for everything f12n
	Object map[string]interface{}
)

// New returns an object from a map
func New() Object {
	return Object{}
}

// FromMap returns an object from a map
func FromMap(m map[string]interface{}) Object {
	o := Object{}
	for k, v := range m {
		o.Set(k, v)
	}
	return o
}

// FromMap inits the object from a map
func (o *Object) FromMap(m map[string]interface{}) error {
	v := FromMap(m)
	*o = v
	return nil
}

// ToObject simply returns a copy of the object
// This is mostly a hack for generated objects
func (o Object) ToObject() Object {
	return o.Copy()
}

// ToMap returns the object as a map
func (o Object) ToMap() map[string]interface{} {
	return map[string]interface{}(o)
}

// GetType returns the object's type
func (o Object) GetType() string {
	if v, ok := o.Get("@type:s").(string); ok {
		return v
	}
	return ""
}

// SetType sets the object's type
func (o Object) SetType(v string) {
	o.Set("@type:s", v)
}

// GetSignature returns the object's signature, or nil
func (o Object) GetSignature() *Object {
	return o.getObject("_signature:o")
}

// SetSignature sets the object's signature
func (o Object) SetSignature(v Object) {
	o.Set("_signature:o", v)
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

// Get -
func (o Object) Get(lk string) interface{} {
	return map[string]interface{}(o)[lk]
}

// Set -
func (o Object) Set(k string, v interface{}) {
	switch ov := v.(type) {
	case Object:
		v = ov.ToMap()
	case []Object:
		os := make([]Object, len(ov))
		for i, o := range ov {
			os[i] = o.ToMap()
		}
		v = os
	}
	map[string]interface{}(o)[k] = v
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

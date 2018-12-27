// Code generated by nimona.io/go/cmd/objectify. DO NOT EDIT.

// +build !generate

package encoding

// ToMap returns a map compatible with f12n
func (s Policy) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"@ctx:s":        "/policy",
		"description:s": s.Description,
		"effect:s":      s.Effect,
	}
	if s.Subjects != nil {
		m["subjects:a<s>"] = s.Subjects
	}
	if s.Actions != nil {
		m["actions:a<s>"] = s.Actions
	}
	return m
}

// ToObject returns a f12n object
func (s Policy) ToObject() *Object {
	return NewObjectFromMap(s.ToMap())
}

// FromMap populates the struct from a f12n compatible map
func (s *Policy) FromMap(m map[string]interface{}) error {
	if v, ok := m["description:s"].(string); ok {
		s.Description = v
	}
	s.Subjects = []string{}
	if ss, ok := m["subjects:a<s>"].([]interface{}); ok {
		for _, si := range ss {
			if v, ok := si.(string); ok {
				s.Subjects = append(s.Subjects, v)
			}
		}
	}
	if v, ok := m["subjects:a<s>"].([]string); ok {
		s.Subjects = v
	}
	s.Actions = []string{}
	if ss, ok := m["actions:a<s>"].([]interface{}); ok {
		for _, si := range ss {
			if v, ok := si.(string); ok {
				s.Actions = append(s.Actions, v)
			}
		}
	}
	if v, ok := m["actions:a<s>"].([]string); ok {
		s.Actions = v
	}
	if v, ok := m["effect:s"].(string); ok {
		s.Effect = v
	}
	return nil
}

// FromObject populates the struct from a f12n object
func (s *Policy) FromObject(o *Object) error {
	return s.FromMap(o.ToMap())
}

// GetType returns the object's type
func (s Policy) GetType() string {
	return "/policy"
}

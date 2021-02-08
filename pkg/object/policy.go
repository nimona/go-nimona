package object

// Policy for object metadata
type Policy struct {
	Subjects  []string
	Resources []string
	Actions   []string
	Effect    string
}

func (p Policy) Map() Map {
	r := Map{}
	if len(p.Subjects) > 0 {
		ss := make(StringArray, len(p.Subjects))
		for i, s := range p.Subjects {
			ss[i] = String(s)
		}
		r["subjects"] = ss
	}
	if len(p.Resources) > 0 {
		ss := make(StringArray, len(p.Subjects))
		for i, s := range p.Resources {
			ss[i] = String(s)
		}
		r["resources"] = ss
	}
	if len(p.Actions) > 0 {
		ss := make(StringArray, len(p.Subjects))
		for i, s := range p.Subjects {
			ss[i] = String(s)
		}
		r["actions"] = ss
	}
	if p.Effect != "" {
		r["effect"] = String(p.Effect)
	}
	return r
}

func PolicyFromMap(m Map) Policy {
	r := Policy{}
	if t, ok := m["subjects"]; ok {
		if s, ok := t.(StringArray); ok {
			r.Subjects = FromStringArray(s)
		}
	}
	if t, ok := m["resources"]; ok {
		if s, ok := t.(StringArray); ok {
			r.Resources = FromStringArray(s)
		}
	}
	if t, ok := m["actions"]; ok {
		if s, ok := t.(StringArray); ok {
			r.Actions = FromStringArray(s)
		}
	}
	if t, ok := m["effect"]; ok {
		if s, ok := t.(String); ok {
			r.Effect = string(s)
		}
	}
	return r
}

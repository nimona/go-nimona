package tilde

func (m Map) Fluent() *FluentMap {
	return &FluentMap{
		Map: m,
	}
}

type FluentMap struct {
	Map
}

func (m *FluentMap) Set(key string, value Value) *FluentMap {
	m.Map.Set(key, value) // TODO: error handling?
	return m
}

func (m *FluentMap) Append(key string, value Value) *FluentMap {
	m.Map.Append(key, value) // TODO: error handling?
	return m
}

func (m *FluentMap) Get(key string) *FluentValue {
	v, err := m.Map.Get(key)
	if err != nil {
		return nil
	}

	return &FluentValue{
		Value: v,
	}
}

type FluentValue struct {
	Value
}

func (v *FluentValue) String() String {
	if v == nil {
		return ""
	}

	s, ok := v.Value.(String)
	if !ok {
		return ""
	}

	return s
}

func (v *FluentValue) Map() Map {
	if v == nil {
		return nil
	}

	m, ok := v.Value.(Map)
	if !ok {
		return nil
	}

	return m
}

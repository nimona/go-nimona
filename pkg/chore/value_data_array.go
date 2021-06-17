package chore

func (v DataArray) Hint() Hint {
	return DataArrayHint
}

func (v DataArray) _isValue() {
}

func (v DataArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}

func (v DataArray) _isArray() {}

func (v DataArray) Len() int {
	return len(v)
}

func (v DataArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

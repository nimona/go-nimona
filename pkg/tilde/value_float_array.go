package tilde

func (v FloatArray) Hint() Hint {
	return FloatArrayHint
}

func (v FloatArray) _isValue() {
}

func (v FloatArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}

func (v FloatArray) _isArray() {}

func (v FloatArray) Len() int {
	return len(v)
}

func (v FloatArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

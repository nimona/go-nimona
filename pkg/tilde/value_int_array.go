package tilde

func (v IntArray) Hint() Hint {
	return IntArrayHint
}

func (v IntArray) _isValue() {
}

func (v IntArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}

func (v IntArray) _isArray() {}

func (v IntArray) Len() int {
	return len(v)
}

func (v IntArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

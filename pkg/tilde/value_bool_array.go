package tilde

func (v BoolArray) Hint() Hint {
	return BoolArrayHint
}

func (v BoolArray) _isValue() {
}

func (v BoolArray) Hash() Digest {
	if v.Len() == 0 {
		return EmptyDigest
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}

func (v BoolArray) _isArray() {
}

func (v BoolArray) Len() int {
	return len(v)
}

func (v BoolArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

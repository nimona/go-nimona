package tilde

func (v StringArray) Hint() Hint {
	return StringArrayHint
}

func (v StringArray) _isValue() {
}

func (v StringArray) Hash() Digest {
	if v.Len() == 0 {
		return EmptyDigest
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}

func (v StringArray) _isArray() {}

func (v StringArray) Len() int {
	return len(v)
}

func (v StringArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

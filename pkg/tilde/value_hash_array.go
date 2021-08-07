package tilde

func (v DigestArray) Hint() Hint {
	return DigestArrayHint
}

func (v DigestArray) _isValue() {
}

func (v DigestArray) Hash() Digest {
	if v.Len() == 0 {
		return EmptyDigest
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}

func (v DigestArray) _isArray() {}

func (v DigestArray) Len() int {
	return len(v)
}

func (v DigestArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

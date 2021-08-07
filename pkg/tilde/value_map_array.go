package tilde

func (v MapArray) Hint() Hint {
	return MapArrayHint
}

func (v MapArray) _isValue() {
}

func (v MapArray) Hash() Digest {
	if v.Len() == 0 {
		return EmptyDigest
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}

func (v MapArray) _isArray() {}

func (v MapArray) Len() int {
	return len(v)
}

func (v MapArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

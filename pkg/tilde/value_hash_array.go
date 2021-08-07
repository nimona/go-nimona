package tilde

func (v HashArray) Hint() Hint {
	return HashArrayHint
}

func (v HashArray) _isValue() {
}

func (v HashArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}

func (v HashArray) _isArray() {}

func (v HashArray) Len() int {
	return len(v)
}

func (v HashArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

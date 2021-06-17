package chore

func (v UintArray) Hint() Hint {
	return UintArrayHint
}

func (v UintArray) _isValue() {
}

func (v UintArray) Hash() Hash {
	if v.Len() == 0 {
		return EmptyHash
	}
	h := []byte{}
	for _, iv := range v {
		h = append(h, iv.Hash()...)
	}
	return hashFromBytes(h)
}
func (v UintArray) _isArray() {}

func (v UintArray) Len() int {
	return len(v)
}

func (v UintArray) Range(f func(int, Value) bool) {
	for k, v := range v {
		if f(k, v) {
			return
		}
	}
}

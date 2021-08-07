package tilde

func (v String) Hint() Hint {
	return StringHint
}

func (v String) _isValue() {
}

func (v String) Hash() Hash {
	if string(v) == "" {
		return EmptyHash
	}
	return hashFromBytes(
		[]byte(string(v)),
	)
}

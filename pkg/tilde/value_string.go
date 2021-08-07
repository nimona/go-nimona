package tilde

func (v String) Hint() Hint {
	return StringHint
}

func (v String) _isValue() {
}

func (v String) Hash() Digest {
	if string(v) == "" {
		return EmptyDigest
	}
	return hashFromBytes(
		[]byte(string(v)),
	)
}

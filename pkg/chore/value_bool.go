package chore

import "encoding/json"

func (v Bool) Hint() Hint {
	return BoolHint
}

func (v Bool) _isValue() {
}

func (v Bool) Hash() Hash {
	if !v {
		return hashFromBytes([]byte{0})
	}
	return hashFromBytes([]byte{1})
}

func (v *Bool) UnmarshalJSON(b []byte) error {
	var iv bool
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Bool(iv)
	return nil
}

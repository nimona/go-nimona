package tilde

import (
	"encoding/json"
	"fmt"
)

func (v Uint) Hint() Hint {
	return UintHint
}

func (v Uint) _isValue() {
}

func (v Uint) Hash() Hash {
	return hashFromBytes(
		[]byte(
			fmt.Sprintf(
				"%d",
				uint64(v),
			),
		),
	)
}

func (v *Uint) UnmarshalJSON(b []byte) error {
	var iv uint64
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Uint(iv)
	return nil
}

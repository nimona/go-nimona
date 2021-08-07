package tilde

import (
	"encoding/json"
	"fmt"
)

func (v Int) Hint() Hint {
	return IntHint
}

func (v Int) _isValue() {
}

func (v Int) Hash() Hash {
	return hashFromBytes(
		[]byte(
			fmt.Sprintf(
				"%d",
				int64(v),
			),
		),
	)
}

func (v *Int) UnmarshalJSON(b []byte) error {
	var iv int64
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Int(iv)
	return nil
}

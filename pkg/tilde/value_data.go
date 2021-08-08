package tilde

import (
	"encoding/json"
	"fmt"
)

func (v Data) Hint() Hint {
	return DataHint
}

func (v Data) _isValue() {
}

func (v Data) Hash() Digest {
	return hashFromBytes(v)
}

func (v *Data) UnmarshalJSON(b []byte) error {
	iv := []byte{}
	err := json.Unmarshal(b, &iv)
	if err != nil {
		fmt.Println(err)
		return err
	}
	*v = Data(iv)
	return nil
}
// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package object

import (
	json "encoding/json"
)

type (
	Hash struct {
		Algorithm string `json:"algorithm:s,omitempty"`
		D         []byte `json:"d:d,omitempty"`
	}
	Link struct {
		Target *Hash `json:"target:o,omitempty"`
	}
)

func (e *Hash) GetType() string {
	return "nimona.io/Hash"
}

func (e *Hash) ToObject() Object {
	m := map[string]interface{}{
		"@ctx:s": "nimona.io/Hash",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return Object(m)
}

func (e *Hash) FromObject(o Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *Link) GetType() string {
	return "nimona.io/Link"
}

func (e *Link) ToObject() Object {
	m := map[string]interface{}{
		"@ctx:s": "nimona.io/Link",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return Object(m)
}

func (e *Link) FromObject(o Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

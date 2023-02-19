package nimona

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"

	"nimona.io/internal/tilde"
)

type DocumentMap struct {
	m tilde.Map
}

func NewDocumentMap(m tilde.Map) *DocumentMap {
	return &DocumentMap{
		m: m,
	}
}

type DocumentValuer interface {
	DocumentValue(v any) any
}

func (m DocumentMap) Type() string {
	if m.m == nil {
		return ""
	}
	vi, err := m.m.Get("$type")
	if err != nil {
		return ""
	}
	v, ok := vi.(tilde.String)
	if !ok {
		return ""
	}
	return string(v)
}

func (m DocumentMap) DocumentMap() DocumentMap {
	return m
}

func (m DocumentMap) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(m.m)
	if err != nil {
		return nil, fmt.Errorf("error marshaling into json: %w", err)
	}
	return b, nil
}

func (m *DocumentMap) UnmarshalJSON(b []byte) error {
	if m == nil {
		*m = DocumentMap{}
	}
	mm := &tilde.Map{}
	err := json.Unmarshal(b, mm)
	if err != nil {
		return fmt.Errorf("error unmarshaling from json: %w", err)
	}
	m.m = *mm
	return nil
}

func DumpDocumentBytes(b []byte) {
	fmt.Printf("%x\n", b)
}

func DumpDocumentMap(m DocumentMapper) {
	yb, err := yaml.Marshal(m.DocumentMap().m)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(yb))
}

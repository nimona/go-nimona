package nimona

import (
	"fmt"
	reflect "reflect"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v2"
)

type DocumentMap map[string]interface{}

func (m DocumentMap) Type() string {
	if m == nil {
		return ""
	}
	if _, ok := m["$type"]; !ok {
		return ""
	}
	return m["$type"].(string)
}

var cborEncoder = func() cbor.EncMode {
	encOpts := cbor.EncOptions{}
	enc, err := encOpts.EncMode()
	if err != nil {
		panic(err)
	}
	return enc
}()

var cborDecoder = func() cbor.DecMode {
	decOpts := cbor.DecOptions{
		DefaultMapType: reflect.TypeOf(DocumentMap{}),
	}
	dec, err := decOpts.DecMode()
	if err != nil {
		panic(err)
	}
	return dec
}()

func (m DocumentMap) DocumentMap() DocumentMap {
	return m
}

// TODO(geoah): this probably needs testing and errors
func (m DocumentMap) FromDocumentMap(dm DocumentMap) {
	for k, v := range dm {
		m[k] = v
	}
}

func (m DocumentMap) MarshalCBOR() ([]byte, error) {
	b, err := cborEncoder.Marshal(map[string]any(m))
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling: %w", err)
	}
	return b, nil
}

func (m *DocumentMap) UnmarshalCBOR(b []byte) (err error) {
	if m == nil {
		*m = map[string]interface{}{}
	}
	mm := &map[string]any{}
	err = cborDecoder.Unmarshal(b, mm)
	if err != nil {
		return fmt.Errorf("error unmarshaling: %w", err)
	}
	*m = DocumentMap(*mm)
	return nil
}

func DumpDocumentBytes(b []byte) {
	fmt.Printf("%x\n", b)
}

func DumpDocumentMap(m DocumentMapper) {
	yb, err := yaml.Marshal(m.DocumentMap())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(yb))
}

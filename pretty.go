package nimona

import (
	"encoding/json"
	"fmt"
)

func PrettyPrintCbor(b []byte) {
	s := PrettySPrintCbor(b)
	fmt.Println(s)
}

func PrettySPrintCbor(b []byte) string {
	m, err := NewDocumentMapFromCBOR(b)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling cbor, err: %w", err))
	}

	jb, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Errorf("error marshaling to json, err: %w", err))
	}

	return fmt.Sprintf("cbor: %x\njson: %s\n", b, string(jb))
}

func PrettyPrint(v Cborer) {
	s := PrettySPrint(v)
	fmt.Println(s)
}

// PrettySPrint returns a string with the cbor and json representation of the
// given cborer.
// WARNING: Should be used for only for debugging as it WILL panic on error.
func PrettySPrint(v Cborer) string {
	b, err := v.MarshalCBORBytes()
	if err != nil {
		panic(fmt.Errorf("error marshaling to cbor, err: %w", err))
	}

	return PrettySPrintCbor(b)
}

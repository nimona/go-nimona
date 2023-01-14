package nimona

import (
	"encoding/json"
	"fmt"
	reflect "reflect"

	"github.com/fxamacker/cbor/v2"
)

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

	m := map[string]interface{}{}
	dm, err := cbor.DecOptions{
		DefaultMapType: reflect.TypeOf(map[string]interface{}{}),
	}.DecMode()
	if err != nil {
		panic(fmt.Errorf("error creating cbor decoder, err: %w", err))
	}

	err = dm.Unmarshal(b, &m)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling cbor, err: %w", err))
	}

	j, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Errorf("error marshaling to json, err: %w", err))
	}

	return fmt.Sprintf("cbor: %x\njson: %s\n", b, string(j))
}
package nimona

import (
	"fmt"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	"github.com/mitchellh/mapstructure"
)

type MessageWrapper[Wrapped any] struct {
	Type string
	Body Wrapped
}

func (w *MessageWrapper[Wrapped]) FromAny(in MessageWrapper[any]) error {
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "cbor",
		Result:  &w.Body,
	})
	if err != nil {
		return fmt.Errorf("error creating decoder: %w", err)
	}
	err = dec.Decode(in.Body)
	if err != nil {
		return fmt.Errorf("error decoding body: %w", err)
	}

	w.Type = in.Type
	return nil
}

func (w MessageWrapper[Wrapped]) ToAny() MessageWrapper[any] {
	return MessageWrapper[any]{
		Type: w.Type,
		Body: w.Body,
	}
}

func (w MessageWrapper[Wrapped]) MarshalCBOR() ([]byte, error) {
	m := map[string]interface{}{
		"$type": w.Type,
	}
	// go through the fields of the wrapped struct and add them to the map.
	// use reflection to get the field names, values, and the cbor tag.
	// if the cbor tag is empty, use the field name as the key.
	// we only do this on the first level of the struct, it's up to body's
	// MarshalCBOR method to handle nested structs.
	v := reflect.ValueOf(w.Body)
	// if the wrapped struct is nil or zero, return
	if !v.IsValid() {
		return cbor.Marshal(m)
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("cbor")
		if tag == "" {
			tag = field.Name
		}
		m[tag] = v.Field(i).Interface()
	}
	return cbor.Marshal(m)
}

func (w *MessageWrapper[Wrapped]) UnmarshalCBOR(data []byte) error {
	// first unmarshal the type into a temporary struct
	tw := struct {
		Type string `cbor:"$type"`
	}{}
	err := cbor.Unmarshal(data, &tw)
	if err != nil {
		return fmt.Errorf("error unmarshaling type: %w", err)
	}
	// then set the type of the wrapper
	w.Type = tw.Type
	// finally unmarshal the body into the wrapped struct
	err = cbor.Unmarshal(data, &w.Body)
	if err != nil {
		return fmt.Errorf("error unmarshaling body: %w", err)
	}
	return nil
}

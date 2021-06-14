// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package blob

import (
	chore "nimona.io/pkg/chore"
	object "nimona.io/pkg/object"
)

type (
	Chunk struct {
		Metadata object.Metadata `nimona:"@metadata:m"`
		Data     []byte          `nimona:"data:d"`
	}
	Blob struct {
		Metadata object.Metadata `nimona:"@metadata:m"`
		Chunks   []chore.CID     `nimona:"chunks:as"`
	}
)

func (e *Chunk) Type() string {
	return "nimona.io/Chunk"
}

func (e *Chunk) MarshalObject() (*object.Object, error) {
	o, err := object.Marshal(e)
	if err != nil {
		return nil, err
	}
	o.Type = "nimona.io/Chunk"
	return o, nil
}

func (e *Chunk) UnmarshalObject(o *object.Object) error {
	return object.Unmarshal(o, e)
}

func (e *Blob) Type() string {
	return "nimona.io/Blob"
}

func (e *Blob) MarshalObject() (*object.Object, error) {
	o, err := object.Marshal(e)
	if err != nil {
		return nil, err
	}
	o.Type = "nimona.io/Blob"
	return o, nil
}

func (e *Blob) UnmarshalObject(o *object.Object) error {
	return object.Unmarshal(o, e)
}

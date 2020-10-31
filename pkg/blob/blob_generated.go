// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package blob

import (
	object "nimona.io/pkg/object"
)

type (
	Chunk struct {
		Metadata object.Metadata `nimona:"metadata:m,omitempty"`
		Data     []byte          `nimona:"data:d,omitempty"`
	}
	Blob struct {
		Metadata object.Metadata `nimona:"metadata:m,omitempty"`
		Chunks   []*Chunk        `nimona:"chunks:ao,omitempty"`
	}
)

func (e *Chunk) Type() string {
	return "nimona.io/Chunk"
}

func (e Chunk) ToObject() *object.Object {
	o, err := object.Encode(&e)
	if err != nil {
		panic(err)
	}
	return o
}

func (e *Chunk) FromObject(o *object.Object) error {
	return object.Decode(o, e)
}

func (e *Blob) Type() string {
	return "nimona.io/Blob"
}

func (e Blob) ToObject() *object.Object {
	o, err := object.Encode(&e)
	if err != nil {
		panic(err)
	}
	return o
}

func (e *Blob) FromObject(o *object.Object) error {
	return object.Decode(o, e)
}

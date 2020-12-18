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
		Chunks   []object.Hash   `nimona:"chunks:ar,omitempty"`
	}
)

func (e *Chunk) Type() string {
	return "nimona.io/Chunk"
}

func (e Chunk) ToObject() *object.Object {
	r := &object.Object{
		Type:     "nimona.io/Chunk",
		Metadata: e.Metadata,
		Data:     map[string]interface{}{},
	}
	r.Data["data:d"] = e.Data
	return r
}

func (e Chunk) ToObjectMap() map[string]interface{} {
	d := map[string]interface{}{}
	d["data:d"] = e.Data
	r := map[string]interface{}{
		"type:s":     "nimona.io/Chunk",
		"metadata:m": object.MetadataToMap(&e.Metadata),
		"data:m":     d,
	}
	return r
}

func (e *Chunk) FromObject(o *object.Object) error {
	return object.Decode(o, e)
}

func (e *Blob) Type() string {
	return "nimona.io/Blob"
}

func (e Blob) ToObject() *object.Object {
	r := &object.Object{
		Type:     "nimona.io/Blob",
		Metadata: e.Metadata,
		Data:     map[string]interface{}{},
	}
	if len(e.Chunks) > 0 {
		r.Data["chunks:ar"] = e.Chunks
	}
	return r
}

func (e Blob) ToObjectMap() map[string]interface{} {
	d := map[string]interface{}{}
	if len(e.Chunks) > 0 {
		d["chunks:ar"] = e.Chunks
	}
	r := map[string]interface{}{
		"type:s":     "nimona.io/Blob",
		"metadata:m": object.MetadataToMap(&e.Metadata),
		"data:m":     d,
	}
	return r
}

func (e *Blob) FromObject(o *object.Object) error {
	return object.Decode(o, e)
}

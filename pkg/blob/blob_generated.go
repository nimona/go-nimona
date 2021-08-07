// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package blob

import (
	object "nimona.io/pkg/object"
	tilde "nimona.io/pkg/tilde"
)

const ChunkType = "nimona.io/Chunk"

type Chunk struct {
	Metadata object.Metadata `nimona:"@metadata:m,type=nimona.io/Chunk"`
	Data     []byte          `nimona:"data:d"`
}

const BlobType = "nimona.io/Blob"

type Blob struct {
	Metadata object.Metadata `nimona:"@metadata:m,type=nimona.io/Blob"`
	Chunks   []tilde.Digest  `nimona:"chunks:as"`
}

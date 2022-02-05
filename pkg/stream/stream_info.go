package stream

import (
	"nimona.io/pkg/object"
	"nimona.io/pkg/tilde"
)

type (
	ObjectInfo struct {
		Type     string
		Digest   tilde.Digest
		Metadata object.Metadata
	}
	StreamInfo struct {
		RootType   string
		RootDigest tilde.Digest
		RootObject *object.Object
		Objects    map[tilde.Digest]*ObjectInfo
	}
)

func GetObjectInfo(o *object.Object) *ObjectInfo {
	return &ObjectInfo{
		Type:     o.Type,
		Digest:   o.Hash(),
		Metadata: o.Metadata,
	}
}

func NewStreamInfo() *StreamInfo {
	return &StreamInfo{
		Objects: map[tilde.Digest]*ObjectInfo{},
	}
}

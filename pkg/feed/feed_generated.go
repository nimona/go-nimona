// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package feed

import (
	object "nimona.io/pkg/object"
	tilde "nimona.io/pkg/tilde"
)

const FeedStreamRootType = "stream:nimona.io/feed"

type FeedStreamRoot struct {
	Metadata   object.Metadata `nimona:"@metadata:m,type=stream:nimona.io/feed"`
	ObjectType string          `nimona:"objectType:s"`
	Timestamp  string          `nimona:"timestamp:s"`
}

const AddedType = "event:nimona.io/feed.Added"

type Added struct {
	Metadata   object.Metadata `nimona:"@metadata:m,type=event:nimona.io/feed.Added"`
	ObjectHash []tilde.Digest  `nimona:"objectHash:ar"`
	Sequence   int64           `nimona:"sequence:i"`
	Timestamp  string          `nimona:"timestamp:s"`
}

const RemovedType = "event:nimona.io/feed.Removed"

type Removed struct {
	Metadata   object.Metadata `nimona:"@metadata:m,type=event:nimona.io/feed.Removed"`
	ObjectHash []tilde.Digest  `nimona:"objectHash:ar"`
	Sequence   int64           `nimona:"sequence:i"`
	Timestamp  string          `nimona:"timestamp:s"`
}

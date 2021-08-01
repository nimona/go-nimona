// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package relationship

import (
	crypto "nimona.io/pkg/crypto"
	object "nimona.io/pkg/object"
)

const RelationshipStreamRootType = "stream:nimona.io/schema/relationship"

type RelationshipStreamRoot struct {
	Metadata object.Metadata `nimona:"@metadata:m,type=stream:nimona.io/schema/relationship"`
}

const AddedType = "event:nimona.io/schema/relationship.Added"

type Added struct {
	Metadata    object.Metadata  `nimona:"@metadata:m,type=event:nimona.io/schema/relationship.Added"`
	Alias       string           `nimona:"alias:s"`
	RemoteParty crypto.PublicKey `nimona:"remoteParty:s"`
	Timestamp   string           `nimona:"timestamp:s"`
}

const RemovedType = "event:nimona.io/schema/relationship.Removed"

type Removed struct {
	Metadata    object.Metadata  `nimona:"@metadata:m,type=event:nimona.io/schema/relationship.Removed"`
	RemoteParty crypto.PublicKey `nimona:"remoteParty:s"`
	Timestamp   string           `nimona:"timestamp:s"`
}

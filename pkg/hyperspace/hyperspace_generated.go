// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package hyperspace

import (
	object "nimona.io/pkg/object"
	peer "nimona.io/pkg/peer"
	tilde "nimona.io/pkg/tilde"
)

const AnnouncementType = "nimona.io/hyperspace.Announcement"

type Announcement struct {
	Metadata         object.Metadata      `nimona:"@metadata:m,type=nimona.io/hyperspace.Announcement"`
	Version          int64                `nimona:"version:i"`
	ConnectionInfo   *peer.ConnectionInfo `nimona:"connectionInfo:m"`
	PeerCapabilities []string             `nimona:"peerCapabilities:as"`
	Digests          []tilde.Digest       `nimona:"digests:ar"`
}

const LookupByDIDRequestType = "nimona.io/hyperspace.LookupByDIDRequest"

type LookupByDIDRequest struct {
	Metadata            object.Metadata `nimona:"@metadata:m,type=nimona.io/hyperspace.LookupByDIDRequest"`
	RequestID           string          `nimona:"requestID:s"`
	Owner               peer.ID         `nimona:"owner:s"`
	RequireCapabilities []string        `nimona:"requireCapabilities:as"`
}

const LookupByDigestRequestType = "nimona.io/hyperspace.LookupByDigestRequest"

type LookupByDigestRequest struct {
	Metadata  object.Metadata `nimona:"@metadata:m,type=nimona.io/hyperspace.LookupByDigestRequest"`
	RequestID string          `nimona:"requestID:s"`
	Digest    tilde.Digest    `nimona:"digest:r"`
}

const LookupResponseType = "nimona.io/hyperspace.LookupResponse"

type LookupResponse struct {
	Metadata      object.Metadata `nimona:"@metadata:m,type=nimona.io/hyperspace.LookupResponse"`
	RequestID     string          `nimona:"requestID:s"`
	Announcements []*Announcement `nimona:"announcements:am"`
}

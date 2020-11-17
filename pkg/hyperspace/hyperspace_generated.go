// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package hyperspace

import (
	object "nimona.io/pkg/object"
	peer "nimona.io/pkg/peer"
)

type (
	Announcement struct {
		Metadata         object.Metadata      `nimona:"metadata:m,omitempty"`
		Version          int64                `nimona:"version:i,omitempty"`
		ConnectionInfo   *peer.ConnectionInfo `nimona:"connectionInfo:o,omitempty"`
		PeerVector       []uint64             `nimona:"peerVector:au,omitempty"`
		PeerCapabilities []string             `nimona:"peerCapabilities:as,omitempty"`
	}
	LookupRequest struct {
		Metadata            object.Metadata `nimona:"metadata:m,omitempty"`
		Nonce               string          `nimona:"nonce:s,omitempty"`
		QueryVector         []uint64        `nimona:"queryVector:au,omitempty"`
		RequireCapabilities []string        `nimona:"requireCapabilities:as,omitempty"`
	}
	LookupResponse struct {
		Metadata      object.Metadata `nimona:"metadata:m,omitempty"`
		Nonce         string          `nimona:"nonce:s,omitempty"`
		QueryVector   []uint64        `nimona:"queryVector:au,omitempty"`
		Announcements []*Announcement `nimona:"announcements:ao,omitempty"`
	}
)

func (e *Announcement) Type() string {
	return "nimona.io/hyperspace.Announcement"
}

func (e Announcement) ToObject() *object.Object {
	r := &object.Object{
		Type:     "nimona.io/hyperspace.Announcement",
		Metadata: e.Metadata,
		Data:     map[string]interface{}{},
	}
	r.Data["version:i"] = e.Version
	if e.ConnectionInfo != nil {
		r.Data["connectionInfo:o"] = e.ConnectionInfo.ToObject()
	}
	if len(e.PeerVector) > 0 {
		// rv := make([]uint64, len(e.PeerVector))
		// for i, v := range e.PeerVector {
		// 	rv[i] = v
		// }
		r.Data["peerVector:au"] = e.PeerVector
	}
	if len(e.PeerCapabilities) > 0 {
		// rv := make([]string, len(e.PeerCapabilities))
		// for i, v := range e.PeerCapabilities {
		// 	rv[i] = v
		// }
		r.Data["peerCapabilities:as"] = e.PeerCapabilities
	}
	return r
}

func (e *Announcement) FromObject(o *object.Object) error {
	return object.Decode(o, e)
}

func (e *LookupRequest) Type() string {
	return "nimona.io/hyperspace.LookupRequest"
}

func (e LookupRequest) ToObject() *object.Object {
	r := &object.Object{
		Type:     "nimona.io/hyperspace.LookupRequest",
		Metadata: e.Metadata,
		Data:     map[string]interface{}{},
	}
	r.Data["nonce:s"] = e.Nonce
	if len(e.QueryVector) > 0 {
		// rv := make([]uint64, len(e.QueryVector))
		// for i, v := range e.QueryVector {
		// 	rv[i] = v
		// }
		r.Data["queryVector:au"] = e.QueryVector
	}
	if len(e.RequireCapabilities) > 0 {
		// rv := make([]string, len(e.RequireCapabilities))
		// for i, v := range e.RequireCapabilities {
		// 	rv[i] = v
		// }
		r.Data["requireCapabilities:as"] = e.RequireCapabilities
	}
	return r
}

func (e *LookupRequest) FromObject(o *object.Object) error {
	return object.Decode(o, e)
}

func (e *LookupResponse) Type() string {
	return "nimona.io/hyperspace.LookupResponse"
}

func (e LookupResponse) ToObject() *object.Object {
	r := &object.Object{
		Type:     "nimona.io/hyperspace.LookupResponse",
		Metadata: e.Metadata,
		Data:     map[string]interface{}{},
	}
	r.Data["nonce:s"] = e.Nonce
	if len(e.QueryVector) > 0 {
		// rv := make([]uint64, len(e.QueryVector))
		// for i, v := range e.QueryVector {
		// 	rv[i] = v
		// }
		r.Data["queryVector:au"] = e.QueryVector
	}
	if len(e.Announcements) > 0 {
		rv := make([]*object.Object, len(e.Announcements))
		for i, v := range e.Announcements {
			rv[i] = v.ToObject()
		}
		r.Data["announcements:ao"] = rv
	}
	return r
}

func (e *LookupResponse) FromObject(o *object.Object) error {
	return object.Decode(o, e)
}

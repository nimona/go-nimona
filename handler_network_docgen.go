// Code generated by nimona.io. DO NOT EDIT.

package nimona

import (
	"github.com/vikyd/zero"
)

var _ = zero.IsZeroVal

func (t *NetworkAnnouncePeerRequest) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "core/network/announcePeer.request"
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.Metadata) {
			m["$metadata"] = t.Metadata.DocumentMap()
		}
	}

	// # t.PeerInfo
	//
	// Type: nimona.PeerInfo, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.PeerInfo) {
			m["peerInfo"] = t.PeerInfo.DocumentMap()
		}
	}

	return m
}

func (t *NetworkAnnouncePeerRequest) FromDocumentMap(m DocumentMap) {
	*t = NetworkAnnouncePeerRequest{}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, ok := m["$metadata"].(DocumentMap); ok {
			e := Metadata{}
			e.FromDocumentMap(v)
			t.Metadata = e
		}
	}

	// # t.PeerInfo
	//
	// Type: nimona.PeerInfo, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, ok := m["peerInfo"].(DocumentMap); ok {
			e := PeerInfo{}
			e.FromDocumentMap(v)
			t.PeerInfo = e
		}
	}

}
func (t *NetworkAnnouncePeerResponse) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "core/network/announcePeer.response"
	}

	// # t.Error
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Error) {
			m["error"] = t.Error
		}
	}

	// # t.ErrorDescription
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.ErrorDescription) {
			m["errorDescription"] = t.ErrorDescription
		}
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.Metadata) {
			m["$metadata"] = t.Metadata.DocumentMap()
		}
	}

	return m
}

func (t *NetworkAnnouncePeerResponse) FromDocumentMap(m DocumentMap) {
	*t = NetworkAnnouncePeerResponse{}

	// # t.Error
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["error"].(bool); ok {
			t.Error = v
		}
	}

	// # t.ErrorDescription
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["errorDescription"].(string); ok {
			t.ErrorDescription = v
		}
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, ok := m["$metadata"].(DocumentMap); ok {
			e := Metadata{}
			e.FromDocumentMap(v)
			t.Metadata = e
		}
	}

}
func (t *NetworkInfoRequest) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "core/network/info.request"
	}

	return m
}

func (t *NetworkInfoRequest) FromDocumentMap(m DocumentMap) {
	*t = NetworkInfoRequest{}

}
func (t *NetworkJoinRequest) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "core/network/join.request"
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.Metadata) {
			m["$metadata"] = t.Metadata.DocumentMap()
		}
	}

	// # t.RequestedHandle
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.RequestedHandle) {
			m["requestedHandle"] = t.RequestedHandle
		}
	}

	return m
}

func (t *NetworkJoinRequest) FromDocumentMap(m DocumentMap) {
	*t = NetworkJoinRequest{}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, ok := m["$metadata"].(DocumentMap); ok {
			e := Metadata{}
			e.FromDocumentMap(v)
			t.Metadata = e
		}
	}

	// # t.RequestedHandle
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["requestedHandle"].(string); ok {
			t.RequestedHandle = v
		}
	}

}
func (t *NetworkJoinResponse) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "core/network/join.response"
	}

	// # t.Accepted
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["accepted"] = t.Accepted
	}

	// # t.Error
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Error) {
			m["error"] = t.Error
		}
	}

	// # t.ErrorDescription
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.ErrorDescription) {
			m["errorDescription"] = t.ErrorDescription
		}
	}

	// # t.Handle
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Handle) {
			m["handle"] = t.Handle
		}
	}

	return m
}

func (t *NetworkJoinResponse) FromDocumentMap(m DocumentMap) {
	*t = NetworkJoinResponse{}

	// # t.Accepted
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["accepted"].(bool); ok {
			t.Accepted = v
		}
	}

	// # t.Error
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["error"].(bool); ok {
			t.Error = v
		}
	}

	// # t.ErrorDescription
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["errorDescription"].(string); ok {
			t.ErrorDescription = v
		}
	}

	// # t.Handle
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["handle"].(string); ok {
			t.Handle = v
		}
	}

}
func (t *NetworkLookupPeerRequest) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "core/network/lookupPeer.request"
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.Metadata) {
			m["$metadata"] = t.Metadata.DocumentMap()
		}
	}

	// # t.PeerKey
	//
	// Type: nimona.PeerKey, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.PeerKey) {
			m["peerKey"] = t.PeerKey.DocumentMap()
		}
	}

	return m
}

func (t *NetworkLookupPeerRequest) FromDocumentMap(m DocumentMap) {
	*t = NetworkLookupPeerRequest{}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, ok := m["$metadata"].(DocumentMap); ok {
			e := Metadata{}
			e.FromDocumentMap(v)
			t.Metadata = e
		}
	}

	// # t.PeerKey
	//
	// Type: nimona.PeerKey, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, ok := m["peerKey"].(DocumentMap); ok {
			e := PeerKey{}
			e.FromDocumentMap(v)
			t.PeerKey = e
		}
	}

}
func (t *NetworkLookupPeerResponse) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "core/network/lookupPeer.response"
	}

	// # t.Error
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Error) {
			m["error"] = t.Error
		}
	}

	// # t.ErrorDescription
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.ErrorDescription) {
			m["errorDescription"] = t.ErrorDescription
		}
	}

	// # t.Found
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Found) {
			m["found"] = t.Found
		}
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.Metadata) {
			m["$metadata"] = t.Metadata.DocumentMap()
		}
	}

	// # t.PeerInfo
	//
	// Type: nimona.PeerInfo, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.PeerInfo) {
			m["peerInfo"] = t.PeerInfo.DocumentMap()
		}
	}

	return m
}

func (t *NetworkLookupPeerResponse) FromDocumentMap(m DocumentMap) {
	*t = NetworkLookupPeerResponse{}

	// # t.Error
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["error"].(bool); ok {
			t.Error = v
		}
	}

	// # t.ErrorDescription
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["errorDescription"].(string); ok {
			t.ErrorDescription = v
		}
	}

	// # t.Found
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["found"].(bool); ok {
			t.Found = v
		}
	}

	// # t.Metadata
	//
	// Type: nimona.Metadata, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, ok := m["$metadata"].(DocumentMap); ok {
			e := Metadata{}
			e.FromDocumentMap(v)
			t.Metadata = e
		}
	}

	// # t.PeerInfo
	//
	// Type: nimona.PeerInfo, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, ok := m["peerInfo"].(DocumentMap); ok {
			e := PeerInfo{}
			e.FromDocumentMap(v)
			t.PeerInfo = e
		}
	}

}
func (t *NetworkResolveHandleRequest) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "core/network/resolveHandle.request"
	}

	// # t.Handle
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Handle) {
			m["handle"] = t.Handle
		}
	}

	return m
}

func (t *NetworkResolveHandleRequest) FromDocumentMap(m DocumentMap) {
	*t = NetworkResolveHandleRequest{}

	// # t.Handle
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["handle"].(string); ok {
			t.Handle = v
		}
	}

}
func (t *NetworkResolveHandleResponse) DocumentMap() DocumentMap {
	m := DocumentMap{}

	// # t.$type
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		m["$type"] = "core/network/resolveHandle.response"
	}

	// # t.Error
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Error) {
			m["error"] = t.Error
		}
	}

	// # t.ErrorDescription
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.ErrorDescription) {
			m["errorDescription"] = t.ErrorDescription
		}
	}

	// # t.Found
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if !zero.IsZeroVal(t.Found) {
			m["found"] = t.Found
		}
	}

	// # t.IdentityID
	//
	// Type: nimona.Identity, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if !zero.IsZeroVal(t.IdentityID) {
			m["identityID"] = t.IdentityID.DocumentMap()
		}
	}

	// # t.PeerAddresses
	//
	// Type: []nimona.PeerAddr, Kind: slice
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: nimona.PeerAddr, ElemKind: struct
	// IsElemSlice: false, IsElemStruct: true, IsElemPointer: false
	{
		if !zero.IsZeroVal(t.PeerAddresses) {
			sm := []any{}
			for _, v := range t.PeerAddresses {
				if !zero.IsZeroVal(t.PeerAddresses) {
					sm = append(sm, v.DocumentMap())
				}
			}
			m["peerAddresses"] = sm
		}
	}

	return m
}

func (t *NetworkResolveHandleResponse) FromDocumentMap(m DocumentMap) {
	*t = NetworkResolveHandleResponse{}

	// # t.Error
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["error"].(bool); ok {
			t.Error = v
		}
	}

	// # t.ErrorDescription
	//
	// Type: string, Kind: string
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["errorDescription"].(string); ok {
			t.ErrorDescription = v
		}
	}

	// # t.Found
	//
	// Type: bool, Kind: bool
	// IsSlice: false, IsStruct: false, IsPointer: false
	{
		if v, ok := m["found"].(bool); ok {
			t.Found = v
		}
	}

	// # t.IdentityID
	//
	// Type: nimona.Identity, Kind: struct
	// IsSlice: false, IsStruct: true, IsPointer: false
	{
		if v, ok := m["identityID"].(DocumentMap); ok {
			e := Identity{}
			e.FromDocumentMap(v)
			t.IdentityID = e
		}
	}

	// # t.PeerAddresses
	//
	// Type: []nimona.PeerAddr, Kind: slice
	// IsSlice: true, IsStruct: false, IsPointer: false
	//
	// ElemType: nimona.PeerAddr, ElemKind: struct
	// IsElemSlice: false, IsElemStruct: true, IsElemPointer: false
	{
		sm := []PeerAddr{}
		if vs, ok := m["peerAddresses"].([]any); ok {
			for _, vi := range vs {
				v, ok := vi.(DocumentMap)
				if ok {
					e := PeerAddr{}
					e.FromDocumentMap(v)
					sm = append(sm, e)
				}
			}
		}
		if len(sm) > 0 {
			t.PeerAddresses = sm
		}
	}

}

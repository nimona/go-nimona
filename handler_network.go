package nimona

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	NetworkInfoRequest struct {
		_ string `cborgen:"$type,const=core/network/info.request"`
	}
	NetworkJoinRequest struct {
		_               string   `cborgen:"$type,const=core/network/join.request"`
		Metadata        Metadata `cborgen:"$metadata,omitempty"`
		RequestedHandle string   `cborgen:"requestedHandle,omitempty"`
	}
	NetworkJoinResponse struct {
		_                string `cborgen:"$type,const=core/network/join.response"`
		Handle           string `cborgen:"handle,omitempty"`
		Accepted         bool   `cborgen:"accepted"`
		Error            bool   `cborgen:"error,omitempty"`
		ErrorDescription string `cborgen:"errorDescription,omitempty"`
	}
	NetworkResolveHandleRequest struct {
		_      string `cborgen:"$type,const=core/network/resolveHandle.request"`
		Handle string `cborgen:"handle,omitempty"`
	}
	NetworkResolveHandleResponse struct {
		_                string     `cborgen:"$type,const=core/network/resolveHandle.response"`
		IdentityID       Identity   `cborgen:"identityID,omitempty"`
		PeerAddresses    []PeerAddr `cborgen:"peerAddresses,omitempty"`
		Found            bool       `cborgen:"found,omitempty"`
		Error            bool       `cborgen:"error,omitempty"`
		ErrorDescription string     `cborgen:"errorDescription,omitempty"`
	}
	NetworkAnnouncePeerRequest struct {
		_        string   `cborgen:"$type,const=core/network/announcePeer.request"`
		Metadata Metadata `cborgen:"$metadata,omitempty"`
		PeerInfo PeerInfo `cborgen:"peerInfo,omitempty"`
	}
	NetworkAnnouncePeerResponse struct {
		_                string   `cborgen:"$type,const=core/network/announcePeer.response"`
		Metadata         Metadata `cborgen:"$metadata,omitempty"`
		Error            bool     `cborgen:"error,omitempty"`
		ErrorDescription string   `cborgen:"errorDescription,omitempty"`
	}
	NetworkLookupPeerRequest struct {
		_        string   `cborgen:"$type,const=core/network/lookupPeer.request"`
		Metadata Metadata `cborgen:"$metadata,omitempty"`
		PeerKey  PeerKey  `cborgen:"peerKey,omitempty"`
	}
	NetworkLookupPeerResponse struct {
		_                string   `cborgen:"$type,const=core/network/lookupPeer.response"`
		Metadata         Metadata `cborgen:"$metadata,omitempty"`
		PeerInfo         PeerInfo `cborgen:"peerInfo,omitempty"`
		Found            bool     `cborgen:"found,omitempty"`
		Error            bool     `cborgen:"error,omitempty"`
		ErrorDescription string   `cborgen:"errorDescription,omitempty"`
	}
)

type (
	HandlerNetwork struct {
		Hostname      string
		PeerConfig    *PeerConfig
		PeerAddresses []PeerAddr
		Store         *gorm.DB
	}
	NetworkAccountingModel struct {
		Handle     string    `gorm:"uniqueIndex"`
		IdentityID *Identity `gorm:"primaryKey"`
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}
	NetworkPeerModel struct {
		PeerKey       PeerKey `gorm:"uniqueIndex"`
		PeerInfoBytes []byte
		CreatedAt     time.Time
		UpdatedAt     time.Time
	}
)

func RequestNetworkInfo(
	ctx context.Context,
	ses *Session,
) (*NetworkInfo, error) {
	req := &NetworkInfoRequest{}
	res := &NetworkInfo{}
	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = msgRes.Decode(res)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerNetwork) HandleNetworkInfoRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := &NetworkInfoRequest{}
	err := msg.Decode(req)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}
	res := &NetworkInfo{
		NetworkAlias: NetworkAlias{
			Hostname: h.Hostname,
		},
		PeerAddresses: h.PeerAddresses,
	}
	err = msg.Respond(res)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}

func RequestNetworkJoin(
	ctx context.Context,
	ses *Session,
	peerConfig *PeerConfig,
	requestedHandle string,
) (*NetworkJoinResponse, error) {
	req := &NetworkJoinRequest{
		RequestedHandle: requestedHandle,
	}
	req.Metadata.Owner = peerConfig.GetIdentity()
	if req.Metadata.Owner == nil {
		return nil, fmt.Errorf("cannot join a network without an identity")
	}
	res := &NetworkJoinResponse{}
	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = msgRes.Decode(res)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerNetwork) HandleNetworkJoinRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := &NetworkJoinRequest{}
	err := msg.Decode(req)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}

	respondWithError := func(desc string) error {
		res := &NetworkJoinResponse{
			Error:            true,
			ErrorDescription: desc,
		}
		err = msg.Respond(res)
		if err != nil {
			return fmt.Errorf("error replying: %w", err)
		}
		return nil
	}

	if req.Metadata.Owner == nil {
		return respondWithError("no owner specified")
	}

	if req.RequestedHandle == "" {
		return respondWithError("no handle specified 1")
	}

	acc := &NetworkAccountingModel{}
	que := h.Store.First(acc, "identity_id = ?", req.Metadata.Owner)
	if que.Error == nil {
		return respondWithError("already joined")
	} else if errors.Is(que.Error, gorm.ErrRecordNotFound) {
		// all ok
	} else if que.Error != nil {
		return respondWithError("temporary error, try again later")
	}

	que = h.Store.First(acc, "handle = ?", req.RequestedHandle)
	if que.Error == nil {
		return respondWithError("handle already taken")
	} else if errors.Is(que.Error, gorm.ErrRecordNotFound) {
		// all ok
	} else if que.Error != nil {
		return respondWithError("temporary error, try again later")
	}

	acc.Handle = req.RequestedHandle
	acc.IdentityID = req.Metadata.Owner

	que = h.Store.Create(acc)
	if que.Error != nil {
		return respondWithError("temporary error, try again later")
	}

	res := &NetworkJoinResponse{
		Accepted: true,
		Handle:   req.RequestedHandle,
	}
	err = msg.Respond(res)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}

func RequestNetworkResolveHandle(
	ctx context.Context,
	ses *Session,
	handle string,
) (*NetworkResolveHandleResponse, error) {
	if handle == "" {
		return nil, fmt.Errorf("no handle specified")
	}
	req := &NetworkResolveHandleRequest{
		Handle: handle,
	}
	res := &NetworkResolveHandleResponse{}
	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = msgRes.Decode(res)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerNetwork) HandleNetworkResolveHandleRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := &NetworkResolveHandleRequest{}
	err := msg.Decode(req)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}

	respondWithError := func(desc string) error {
		res := &NetworkResolveHandleResponse{
			Error:            true,
			ErrorDescription: desc,
		}
		err = msg.Respond(res)
		if err != nil {
			return fmt.Errorf("error replying: %w", err)
		}
		return nil
	}

	if req.Handle == "" {
		return respondWithError("no handle specified 3")
	}

	acc := &NetworkAccountingModel{}
	que := h.Store.First(acc, "handle = ?", req.Handle)
	if que.Error == nil {
		// all ok
	} else if errors.Is(que.Error, gorm.ErrRecordNotFound) {
		return respondWithError("not found")
	} else if que.Error != nil {
		return respondWithError("temporary error, try again later")
	}

	res := &NetworkResolveHandleResponse{
		Found:      true,
		IdentityID: *acc.IdentityID,
	}
	err = msg.Respond(res)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}

func RequestNetworkAnnouncePeer(
	ctx context.Context,
	ses *Session,
	peerConfig *PeerConfig,
) (*NetworkAnnouncePeerResponse, error) {
	req := &NetworkAnnouncePeerRequest{
		PeerInfo: *peerConfig.GetPeerInfo(),
	}
	req.Metadata.Owner = peerConfig.GetIdentity()
	if req.Metadata.Owner == nil {
		return nil, fmt.Errorf("cannot announce a peer without an identity")
	}
	res := &NetworkAnnouncePeerResponse{}
	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = msgRes.Decode(res)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerNetwork) HandleNetworkAnnouncePeerRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := &NetworkAnnouncePeerRequest{}
	err := msg.Decode(req)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}

	respondWithError := func(desc string) error {
		res := &NetworkAnnouncePeerResponse{
			Error:            true,
			ErrorDescription: desc,
		}
		err = msg.Respond(res)
		if err != nil {
			return fmt.Errorf("error replying: %w", err)
		}
		return nil
	}

	if req.Metadata.Owner == nil {
		return respondWithError("no owner specified")
	}

	peer := &NetworkPeerModel{
		PeerKey: PeerKey{
			PublicKey: req.PeerInfo.PublicKey,
		},
		PeerInfoBytes: req.PeerInfo.RawBytes,
	}

	cls := clause.OnConflict{UpdateAll: true}
	que := h.Store.Clauses(cls).Create(peer)
	if que.Error != nil {
		return respondWithError("temporary error, try again later")
	}

	res := &NetworkAnnouncePeerResponse{}
	err = msg.Respond(res)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}

func RequestNetworkLookupPeer(
	ctx context.Context,
	ses *Session,
	peerKey PeerKey,
) (*NetworkLookupPeerResponse, error) {
	req := &NetworkLookupPeerRequest{
		PeerKey: peerKey,
	}
	res := &NetworkLookupPeerResponse{}
	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = msgRes.Decode(res)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerNetwork) HandleNetworkLookupPeerRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := &NetworkLookupPeerRequest{}
	err := msg.Decode(req)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}

	respondWithError := func(desc string) error {
		res := &NetworkLookupPeerResponse{
			Error:            true,
			ErrorDescription: desc,
		}
		err = msg.Respond(res)
		if err != nil {
			return fmt.Errorf("error replying: %w", err)
		}
		return nil
	}

	if req.PeerKey.PublicKey.IsZero() {
		return respondWithError("no public key specified")
	}

	entry := &NetworkPeerModel{}
	que := h.Store.First(entry, "peer_key = ?", req.PeerKey.String())
	if que.Error == nil {
		// all ok
	} else if errors.Is(que.Error, gorm.ErrRecordNotFound) {
		return respondWithError("not found")
	} else if que.Error != nil {
		// TODO: log error
		fmt.Println("error looking up peer:", que.Error)
		return respondWithError("temporary error, try again later")
	}

	peerInfo := &PeerInfo{}
	err = UnmarshalCBORBytes(entry.PeerInfoBytes, peerInfo)
	if err != nil {
		// TODO: log error
		fmt.Println("error unmarshaling peer info:", err)
		return respondWithError("temporary error, try again later")
	}

	res := &NetworkLookupPeerResponse{
		Found:    true,
		PeerInfo: *peerInfo,
	}
	err = msg.Respond(res)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}

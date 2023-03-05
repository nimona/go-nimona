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
		_ string `nimona:"$type,type=core/network/info.request"`
	}
	NetworkJoinRequest struct {
		_               string   `nimona:"$type,type=core/network/join.request"`
		Metadata        Metadata `nimona:"$metadata,omitempty"`
		RequestedHandle string   `nimona:"requestedHandle,omitempty"`
	}
	NetworkJoinResponse struct {
		_                string `nimona:"$type,type=core/network/join.response"`
		Handle           string `nimona:"handle,omitempty"`
		Accepted         bool   `nimona:"accepted"`
		Error            bool   `nimona:"error,omitempty"`
		ErrorDescription string `nimona:"errorDescription,omitempty"`
	}
	NetworkResolveHandleRequest struct {
		_      string `nimona:"$type,type=core/network/resolveHandle.request"`
		Handle string `nimona:"handle,omitempty"`
	}
	NetworkResolveHandleResponse struct {
		_                string     `nimona:"$type,type=core/network/resolveHandle.response"`
		IdentityID       Identity   `nimona:"identityID,omitempty"`
		PeerAddresses    []PeerAddr `nimona:"peerAddresses,omitempty"`
		Found            bool       `nimona:"found,omitempty"`
		Error            bool       `nimona:"error,omitempty"`
		ErrorDescription string     `nimona:"errorDescription,omitempty"`
	}
	NetworkAnnouncePeerRequest struct {
		_        string   `nimona:"$type,type=core/network/announcePeer.request"`
		Metadata Metadata `nimona:"$metadata,omitempty"`
		PeerInfo PeerInfo `nimona:"peerInfo,omitempty"`
	}
	NetworkAnnouncePeerResponse struct {
		_                string   `nimona:"$type,type=core/network/announcePeer.response"`
		Metadata         Metadata `nimona:"$metadata,omitempty"`
		Error            bool     `nimona:"error,omitempty"`
		ErrorDescription string   `nimona:"errorDescription,omitempty"`
	}
	NetworkLookupPeerRequest struct {
		_        string   `nimona:"$type,type=core/network/lookupPeer.request"`
		Metadata Metadata `nimona:"$metadata,omitempty"`
		PeerKey  PeerKey  `nimona:"peerKey,omitempty"`
	}
	NetworkLookupPeerResponse struct {
		_                string   `nimona:"$type,type=core/network/lookupPeer.response"`
		Metadata         Metadata `nimona:"$metadata,omitempty"`
		PeerInfo         PeerInfo `nimona:"peerInfo,omitempty"`
		Found            bool     `nimona:"found,omitempty"`
		Error            bool     `nimona:"error,omitempty"`
		ErrorDescription string   `nimona:"errorDescription,omitempty"`
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
	// TODO Res should be a NetworkInfoResponse that contains NetworkInfo or error
	res := &NetworkInfo{}
	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = res.FromDocument(msgRes.Document)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerNetwork) HandleNetworkInfoRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := NetworkInfoRequest{}
	err := req.FromDocument(msg.Document)
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
	err = res.FromDocument(msgRes.Document)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerNetwork) HandleNetworkJoinRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := NetworkJoinRequest{}
	err := req.FromDocument(msg.Document)
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
	err = res.FromDocument(msgRes.Document)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerNetwork) HandleNetworkResolveHandleRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := NetworkResolveHandleRequest{}
	err := req.FromDocument(msg.Document)
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
	err = res.FromDocument(msgRes.Document)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerNetwork) HandleNetworkAnnouncePeerRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := NetworkAnnouncePeerRequest{}
	err := req.FromDocument(msg.Document)
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

	peerInfoBytes, err := req.PeerInfo.Document().MarshalJSON()
	if err != nil {
		return fmt.Errorf("error marshaling peerInfo: %w", err)
	}

	peer := &NetworkPeerModel{
		PeerKey: PeerKey{
			PublicKey: req.PeerInfo.PublicKey,
		},
		PeerInfoBytes: peerInfoBytes,
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
	req := NetworkLookupPeerRequest{
		PeerKey: peerKey,
	}
	res := &NetworkLookupPeerResponse{}
	msgRes, err := ses.Request(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = res.FromDocument(msgRes.Document)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerNetwork) HandleNetworkLookupPeerRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := NetworkLookupPeerRequest{}
	err := req.FromDocument(msg.Document)
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

	peerInfoDoc := &Document{}
	err = peerInfoDoc.UnmarshalJSON(entry.PeerInfoBytes)
	if err != nil {
		// TODO: log error
		fmt.Println("error unmarshaling peer info json into doc:", err)
		return respondWithError("temporary error, try again later")
	}

	peerInfo := &PeerInfo{}
	err = peerInfo.FromDocument(peerInfoDoc)
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

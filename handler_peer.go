package nimona

import (
	"context"
	"fmt"
)

type (
	PeerCapabilitiesRequest struct {
		_ string `nimona:"$type,type=core/peer/capabilities.request"`
	}
	PeerCapabilitiesResponse struct {
		_            string   `nimona:"$type,type=core/peer/capabilities.response"`
		Capabilities []string `cbor:"capabilities"`
	}
)

func RequestPeerCapabilities(
	ctx context.Context,
	ses *Session,
) (*PeerCapabilitiesResponse, error) {
	req := &PeerCapabilitiesRequest{}
	res := &PeerCapabilitiesResponse{}
	msgRes, err := ses.Request(ctx, req.Document())
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = res.FromDocument(msgRes.Document)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func HandlePeerCapabilitiesRequest(
	sesManager *SessionManager,
	capabilities []string,
) {
	handler := func(
		ctx context.Context,
		msg *Request,
	) error {
		req := PeerCapabilitiesRequest{}
		err := req.FromDocument(msg.Document)
		if err != nil {
			return fmt.Errorf("error unmarshaling request: %w", err)
		}
		if msg.Type != "core/peer/capabilities.request" {
			return fmt.Errorf("invalid request type: %s", msg.Type)
		}
		res := &PeerCapabilitiesResponse{
			Capabilities: capabilities,
		}
		err = msg.Respond(res.Document())
		if err != nil {
			return fmt.Errorf("error replying: %w", err)
		}
		return nil
	}
	sesManager.RegisterHandler("core/peer/capabilities.request", handler)
}

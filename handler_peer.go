package nimona

import (
	"context"
	"fmt"
)

type (
	PeerCapabilitiesRequest struct {
		Type string `cborgen:"$type,const=core/peer/capabilities.request"`
	}
	PeerCapabilitiesResponse struct {
		Type         string   `cborgen:"$type,const=core/peer/capabilities.response"`
		Capabilities []string `cbor:"capabilities"`
	}
)

type HandlerPeerCapabilities struct {
	Capabilities []string
}

func RequestPeerCapabilities(
	ctx context.Context,
	ses *Session,
) (*PeerCapabilitiesResponse, error) {
	req := &PeerCapabilitiesRequest{}
	res := &PeerCapabilitiesResponse{}
	err := ses.Request(ctx, req, res)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	return res, nil
}

func (h *HandlerPeerCapabilities) HandlePeerCapabilitiesRequest(
	ctx context.Context,
	msg *MessageRequest,
) error {
	req := &PeerCapabilitiesRequest{}
	err := msg.UnmarsalInto(req)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}
	fmt.Println("Got request", msg)
	if msg.Type != "core/peer/capabilities.request" {
		return fmt.Errorf("invalid request type: %s", msg.Type)
	}
	res := &PeerCapabilitiesResponse{
		Capabilities: h.Capabilities,
	}
	err = msg.Respond(res)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}

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

type HandlerPeerCapabilities struct {
	Capabilities []string
}

func RequestPeerCapabilities(
	ctx context.Context,
	ses *Session,
) (*PeerCapabilitiesResponse, error) {
	req := &PeerCapabilitiesRequest{}
	res := &PeerCapabilitiesResponse{}
	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = msgRes.Codec.Decode(msgRes.Body, res)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerPeerCapabilities) HandlePeerCapabilitiesRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := PeerCapabilitiesRequest{}
	err := msg.Codec.Decode(msg.DocumentRaw, &req)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}
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

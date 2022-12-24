package nimona

import (
	"context"
	"fmt"
)

type (
	PeerCapabilitiesRequest  struct{}
	PeerCapabilitiesResponse struct {
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
	msg := &MessageWrapper[PeerCapabilitiesRequest]{
		Type: "core/peer/capabilities.request",
		Body: PeerCapabilitiesRequest{},
	}
	resAny, err := ses.Request(ctx, msg.ToAny())
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	res := &MessageWrapper[PeerCapabilitiesResponse]{}
	err = res.FromAny(*resAny)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}
	if res.Type != "core/peer/capabilities.response" {
		return nil, fmt.Errorf("invalid response type: %s", res.Type)
	}
	return &res.Body, nil
}

func (h *HandlerPeerCapabilities) HandlePeerCapabilitiesRequest(
	ctx context.Context,
	req *MessageRequest,
) error {
	msg := &MessageWrapper[PeerCapabilitiesRequest]{}
	err := msg.FromAny(req.Body)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}
	fmt.Println("Got request", msg)
	if msg.Type != "core/peer/capabilities.request" {
		return fmt.Errorf("invalid request type: %s", msg.Type)
	}
	res := &MessageWrapper[PeerCapabilitiesResponse]{
		Type: "core/peer/capabilities.response",
		Body: PeerCapabilitiesResponse{
			Capabilities: h.Capabilities,
		},
	}
	err = req.Respond(res.ToAny())
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}

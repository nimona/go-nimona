package nimona

import (
	"context"
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

type (
	PeerCapabilitiesRequest struct {
		Test string `cbor:"test"`
	}
	PeerCapabilitiesResponse struct {
		Capabilities []string `cbor:"capabilities"`
	}
)

type HandlerPeerCapabilities struct {
	Capabilities []string
}

func RequestPeerCapabilities(ctx context.Context, rpc *RPC) (*PeerCapabilitiesResponse, error) {
	msg := &MessageWrapper[PeerCapabilitiesRequest]{
		Type: "core/peer/capabilities.request",
		Body: PeerCapabilitiesRequest{
			Test: "test",
		},
	}
	msgBytes, err := cbor.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("error marshaling message: %w", err)
	}
	res, err := rpc.Request(ctx, msgBytes)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	resMsg := &MessageWrapper[PeerCapabilitiesResponse]{}
	err = cbor.Unmarshal(res, resMsg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}
	if resMsg.Type != "core/peer/capabilities.response" {
		return nil, fmt.Errorf("invalid response type: %s", resMsg.Type)
	}
	return &resMsg.Body, nil
}

func (h *HandlerPeerCapabilities) HandlePeerCapabilitiesRequest(
	ctx context.Context,
	req *Request,
) error {
	msg := &MessageWrapper[PeerCapabilitiesRequest]{}
	err := cbor.Unmarshal(req.Body, msg)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}
	fmt.Println("Got request", msg)
	if msg.Type != "core/peer/capabilities.request" {
		return fmt.Errorf("invalid request type: %s", msg.Type)
	}
	resBody := &MessageWrapper[PeerCapabilitiesResponse]{
		Type: "core/peer/capabilities.response",
		Body: PeerCapabilitiesResponse{
			Capabilities: h.Capabilities,
		},
	}
	resBytes, err := cbor.Marshal(resBody)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}
	err = req.Respond(resBytes)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}

package nimona

import (
	"context"
	"fmt"
)

type (
	NetworkInfoRequest struct {
		Type string `cborgen:"$type,const=core/network/info.request"`
	}
	NetworkInfo struct {
		Type          string     `cborgen:"$type,const=core/network/info"`
		NetworkID     NetworkID  `cborgen:"networkID"`
		PeerAddresses []PeerAddr `cborgen:"peerAddresses"`
	}
)

type HandlerNetwork struct {
	Hostname      string
	PeerAddresses []PeerAddr
}

func RequestNetworkInfo(
	ctx context.Context,
	ses *Session,
) (*NetworkInfo, error) {
	req := &NetworkInfoRequest{}
	res := &NetworkInfo{}
	err := ses.Request(ctx, req, res)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	return res, nil
}

func (h *HandlerNetwork) HandleNetworkInfoRequest(
	ctx context.Context,
	msg *MessageRequest,
) error {
	req := &NetworkInfoRequest{}
	err := msg.UnmarsalInto(req)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}
	res := &NetworkInfo{
		NetworkID: NetworkID{
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

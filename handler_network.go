package nimona

import (
	"context"
	"fmt"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

type (
	NetworkInfoRequest struct {
		_ string `cborgen:"$type,const=core/network/info.request"`
	}
	NetworkInfo struct {
		_             string     `cborgen:"$type,const=core/network/info"`
		Metadata      Metadata   `cborgen:"metadata"`
		NetworkID     NetworkID  `cborgen:"networkID"`
		PeerAddresses []PeerAddr `cborgen:"peerAddresses"`
		RawBytes      []byte     `cborgen:"rawbytes"`
	}
)

type HandlerNetwork struct {
	Hostname      string
	PeerAddresses []PeerAddr
	PrivateKey    ed25519.PrivateKey
}

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

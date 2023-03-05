package nimona

import (
	"context"
	"fmt"
	"time"
)

type (
	Ping struct {
		_     string `nimona:"$type,type=test/ping"`
		Nonce string `nimona:"nonce"`
	}
	Pong struct {
		_     string `nimona:"$type,type=test/pong"`
		Nonce string `nimona:"nonce"`
	}
)

type (
	HandlerPing struct {
		PeerConfig *PeerConfig
	}
)

func RequestPing(
	ctx context.Context,
	ses *Session,
) (*Pong, error) {
	req := &Ping{}
	res := &Pong{}
	msgRes, err := ses.Request(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	err = res.FromDocumentMap(msgRes.Document)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return res, nil
}

func (h *HandlerPing) HandlePingRequest(
	ctx context.Context,
	msg *Request,
) error {
	req := Ping{}
	err := req.FromDocumentMap(msg.Document)
	if err != nil {
		return fmt.Errorf("error unmarshaling request: %w", err)
	}
	res := &Pong{
		Nonce: time.Now().Format(time.RFC3339Nano),
	}
	err = msg.Respond(res)
	if err != nil {
		return fmt.Errorf("error replying: %w", err)
	}
	return nil
}

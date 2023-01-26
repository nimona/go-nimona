package nimona

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type (
	NetworkInfoRequest struct {
		_ string `cborgen:"$type,const=core/network/info.request"`
	}
	NetworkAccountingRegistrationRequest struct {
		_               string   `cborgen:"$type,const=core/network/accounting/registration.request"`
		Metadata        Metadata `cbor:"$metadata,omitempty"`
		RequestedHandle string   `cbor:"requestedHandle,omitempty"`
	}
	NetworkAccountingRegistrationResponse struct {
		_                string `cborgen:"$type,const=core/network/accounting/registration.response"`
		Handle           string `cbor:"handle,omitempty"`
		Accepted         bool   `cbor:"accepted"`
		Error            bool   `cbor:"error,omitempty"`
		ErrorDescription string `cbor:"errorDescription,omitempty"`
	}
)

type (
	HandlerNetwork struct {
		Hostname        string
		PeerAddresses   []PeerAddr
		PrivateKey      PrivateKey
		AccountingStore *gorm.DB
	}
	NetworkAccountingModel struct {
		Handle    string   `gorm:"primaryKey"`
		PeerInfo  PeerInfo `gorm:"embedded"`
		CreatedAt time.Time
		UpdatedAt time.Time
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

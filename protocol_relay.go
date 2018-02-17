package fabric

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
)

type RelayProtocol struct {
}

func (m *RelayProtocol) Name() string {
	return "relay"
}

// The server offering the relay part
func (m *RelayProtocol) Handle(fn HandlerFunc) HandlerFunc {
	return func(ctx context.Context, c Conn) error {
		addr := c.GetAddress()
		lgr := Logger(ctx).With(
			zap.Namespace("protocol:relay"),
			zap.String("addr.current", addr.Current()),
			zap.String("addr.params", addr.CurrentParams()),
		)

		if strings.Contains(addr.CurrentParams(), "keepalive") {
			for {
				token, err := c.ReadToken()
				if err != nil {
					lgr.Error("Could not read token", zap.Error(err))
					return err
				}
				lgr.Debug("Read relay ping", zap.String("token", string(token)))

			}
		}
		return nil
	}
}

// The client that wants to connect
func (m *RelayProtocol) Negotiate(fn NegotiatorFunc) NegotiatorFunc {
	return func(ctx context.Context, c Conn) error {
		addr := c.GetAddress()
		lgr := Logger(ctx).With(
			zap.Namespace("protocol:relay"),
			zap.String("addr.current", addr.Current()),
			zap.String("addr.params", addr.CurrentParams()),
		)

		if strings.Contains(addr.CurrentParams(), "keepalive") {
			for {
				if err := c.WriteToken([]byte("PONG")); err != nil {
					lgr.Error("Could not write token", zap.Error(err))
					return err
				}

				lgr.Debug("Wrote relay ping")
				time.Sleep(5 * time.Second)
			}

		}
		return nil
	}
}

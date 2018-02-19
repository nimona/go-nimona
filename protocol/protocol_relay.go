package protocol

import (
	"context"
	"io"
	"strings"
	"time"

	zap "go.uber.org/zap"

	connection "github.com/nimona/go-nimona-fabric/connection"
	logging "github.com/nimona/go-nimona-fabric/logging"
)

type RelayProtocol struct {
	connections map[string]connection.Conn
	fabric      Dialer
}

func NewRelayProtocol(f Dialer) *RelayProtocol {
	return &RelayProtocol{
		fabric: f,
	}
}
func (m *RelayProtocol) Name() string {
	return "relay"
}

// The server offering the relay part
func (m *RelayProtocol) Handle(fn HandlerFunc) HandlerFunc {
	return func(ctx context.Context, c connection.Conn) error {
		addr := c.GetAddress()
		lgr := logging.Logger(ctx).With(
			zap.Namespace("protocol:relay"),
			zap.String("addr.current", addr.Current()),
			zap.String("addr.params", addr.CurrentParams()),
		)

		lgr.Debug("Handling Relay")

		// For param `keepalive` start listening for a token
		if strings.Contains(addr.CurrentParams(), "keepalive") {
			go func() {
				for {
					_, err := c.ReadToken()
					if err != nil {
						lgr.Error("Could not read token", zap.Error(err))
					}
					// lgr.Debug("Read relay ping", zap.String("token", string(token)))

				}
			}()
			return nil
		}

		host := addr.CurrentParams()
		addr.Pop()

		// Connect
		ctx, cn, err := m.fabric.DialContext(ctx, host)
		if err != nil {
			lgr.Error("Could not Dial client", zap.Error(err))
			return err
		}

		// Pipe
		go func() {
			err = m.pipe(ctx, cn, c)
			if err != nil {
				lgr.Error("Could not start pipe", zap.Error(err))
			}

		}()
		return nil
	}
}

// The client that wants to connect
func (m *RelayProtocol) Negotiate(fn NegotiatorFunc) NegotiatorFunc {
	return func(ctx context.Context, c connection.Conn) error {
		addr := c.GetAddress()
		lgr := logging.Logger(ctx).With(
			zap.Namespace("protocol:relay"),
			zap.String("addr.current", addr.Current()),
			zap.String("addr.params", addr.CurrentParams()),
		)

		lgr.Debug("Negotiating Relay")
		if strings.Contains(addr.CurrentParams(), "keepalive") {
			go func() {
				for {
					time.Sleep(10 * time.Second)
					if err := c.WriteToken([]byte("PONG")); err != nil {
						lgr.Error("Could not write token", zap.Error(err))
					}

					// lgr.Debug("Wrote relay ping")
				}
			}()
			return nil
		}

		addr.Pop()
		return fn(ctx, c)
	}
}

func (m *RelayProtocol) pipe(ctx context.Context,
	a, b io.ReadWriteCloser) error {
	lgr := logging.Logger(ctx).With(
		zap.Namespace("protocol:relay"),
	)

	lgr.Info("Piping")
	done := make(chan error, 1)

	cp := func(r, w io.ReadWriteCloser) {
		n, err := io.Copy(r, w)
		if err != nil {
			lgr.Error("Failed to copy bytes", zap.Error(err))
		}
		lgr.Debug("Bytes written", zap.Int64("bytes", n))
		done <- err
	}
	go cp(a, b)
	go cp(b, a)
	return <-done
}

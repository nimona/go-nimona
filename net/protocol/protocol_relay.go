package protocol

import (
	"context"
	"io"
	"strings"
	"time"

	nnet "github.com/nimona/go-nimona/net"
	zap "go.uber.org/zap"
)

type RelayProtocol struct {
	connections map[string]nnet.Conn
	net         nnet.Net
}

func NewRelayProtocol(f nnet.Net) *RelayProtocol {
	return &RelayProtocol{
		net: f,
	}
}
func (m *RelayProtocol) Name() string {
	return "relay"
}

// The server offering the relay part
func (m *RelayProtocol) Handle(fn nnet.HandlerFunc) nnet.HandlerFunc {
	return func(ctx context.Context, c nnet.Conn) error {
		addr := c.GetAddress()
		lgr := nnet.Logger(ctx).With(
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
		ctx, cn, err := m.net.DialContext(ctx, host)
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
func (m *RelayProtocol) Negotiate(fn nnet.NegotiatorFunc) nnet.NegotiatorFunc {
	return func(ctx context.Context, c nnet.Conn) error {
		addr := c.GetAddress()
		lgr := nnet.Logger(ctx).With(
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
	lgr := nnet.Logger(ctx).With(
		zap.Namespace("protocol:relay"),
	)

	lgr.Debug("Piping")
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
func (s *RelayProtocol) GetAddresses() []string {
	return []string{}
}

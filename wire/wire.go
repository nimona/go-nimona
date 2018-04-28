package wire

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/coreos/go-semver/semver"

	"go.uber.org/zap"

	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/net"
)

type EventHandler func(event *Message) error

const (
	messagingProtocolName      = "wire"
	messagingProtocolVersion   = "0.1.0"
	messagingProtocolCodecJSON = "json"
)

var messagingProtocolVersionCanon = semver.New(messagingProtocolVersion)

type Wire struct {
	mesh     mesh.Mesh
	registry mesh.Registry
	streams  map[string]io.ReadWriteCloser
	handlers map[string]EventHandler
	logger   *zap.Logger
}

func NewWire(ms mesh.Mesh, reg mesh.Registry) (*Wire, error) {
	ctx := context.Background()

	m := &Wire{
		mesh:     ms,
		registry: reg,
		streams:  map[string]io.ReadWriteCloser{},
		handlers: map[string]EventHandler{},
		logger:   net.Logger(ctx).Named("Wire"),
	}

	return m, nil
}

func (m *Wire) HandleExtensionEvents(extension string, h EventHandler) error {
	if _, ok := m.handlers[extension]; ok {
		return errors.New("There already is a handler registered for this extension")
	}
	m.handlers[extension] = h
	return nil
}

// Negotiate will be called after all the other protocol have been processed
func (m *Wire) Negotiate(fn net.NegotiatorFunc) net.NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c net.Conn) error {
		return fn(ctx, c)
	}
}

// Handle adds the base protocols for transports
func (m *Wire) Handle(fn net.HandlerFunc) net.HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c net.Conn) error {
		scanner := bufio.NewScanner(c)
		for scanner.Scan() {
			line := scanner.Text()
			m.Process(ctx, []byte(line))
		}
		return nil
	}
}

func (m *Wire) Process(ctx context.Context, bs []byte) error {
	msg := &Message{}
	if err := json.Unmarshal(bs, &msg); err != nil {
		return err
	}

	hn, ok := m.handlers[msg.Extension]
	if !ok {
		m.logger.Info(
			"No handler registered for extention",
			zap.String("extension", msg.Extension),
		)
		return errors.New("no handler")
	}

	if err := json.Unmarshal(bs, &msg); err != nil {
		m.logger.Info(
			"Could not unmarshal  into type",
			zap.String("extension", msg.Extension),
			zap.String("payload_type", msg.PayloadType),
			zap.Error(err),
		)
		return errors.New("could not unmarshal into type")
	}

	if err := hn(msg); err != nil {
		m.logger.Info(
			"Could not handle event",
			zap.String("extension", msg.Extension),
			zap.String("payload_type", msg.PayloadType),
			zap.Error(err),
		)
		return errors.New("could not handle")
	}

	return nil
}

func (m *Wire) Send(ctx context.Context, extention, payloadType string, payload interface{}, to []string) error {
	if len(to) == 0 {
		return nil
	}

	bs, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	for _, recipient := range to {
		if recipient == "" {
			continue
		}
		event := &Message{
			Version:     *messagingProtocolVersionCanon,
			Codec:       messagingProtocolCodecJSON,
			Extension:   extention,
			PayloadType: payloadType,
			Payload:     bs,
			From:        m.registry.GetLocalPeerInfo().ID,
			To:          recipient,
		}
		if err := m.sendMessage(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (m *Wire) Name() string {
	return messagingProtocolName
}

func (m *Wire) GetAddresses() []string {
	return []string{m.Name()}
}

func (m *Wire) sendMessage(ctx context.Context, msg *Message) error {
	logger := m.logger.With(zap.String("peerID", msg.To))
	stream, ok := m.streams[msg.To]
	if !ok || stream == nil {
		_, conn, err := m.mesh.Dial(ctx, msg.To, messagingProtocolName)
		if err != nil {
			m.logger.Warn("could not dial to peer",
				zap.Error(err),
				zap.String("peerID", msg.To),
				zap.Error(err),
			)
			return err
		}

		m.streams[msg.To] = conn
		stream = conn
		return nil
	}

	logger.Info("Attempting to write message")
	b, err := json.Marshal(msg)
	if err != nil {
		m.logger.Warn("could not marshal outgoing message", zap.Error(err))
		return err
	}

	b = append(b, '\n')
	if _, err := stream.Write(b); err != nil {
		m.logger.Warn("could not write outgoing message", zap.Error(err))
		delete(m.streams, msg.To)
		stream.Close()
		return err
	}

	m.logger.Debug("Wrote message", zap.String("peerID", msg.To))
	return nil
}

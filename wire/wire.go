package wire

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"

	"github.com/coreos/go-semver/semver"
	"go.uber.org/zap"

	"github.com/nimona/go-nimona/mesh"
)

type EventHandler func(event *Message) error

const (
	messagingProtocolName      = "wire"
	messagingProtocolVersion   = "0.1.0"
	messagingProtocolCodecJSON = "json"
)

var messagingProtocolVersionCanon = semver.New(messagingProtocolVersion)

type Wire interface {
	mesh.Handler
	HandleExtensionEvents(extension string, h EventHandler) error
	Send(ctx context.Context, extention, payloadType string, payload interface{}, to []string) error
}

type wire struct {
	mesh     mesh.Mesh
	registry mesh.Registry
	streams  map[string]io.ReadWriteCloser
	handlers map[string]EventHandler
	logger   *zap.Logger
}

func NewWire(ms mesh.Mesh, reg mesh.Registry) (Wire, error) {
	ctx := context.Background()

	m := &wire{
		mesh:     ms,
		registry: reg,
		streams:  map[string]io.ReadWriteCloser{},
		handlers: map[string]EventHandler{},
		logger:   mesh.Logger(ctx).Named("wire"),
	}

	return m, nil
}

func (m *wire) HandleExtensionEvents(extension string, h EventHandler) error {
	if _, ok := m.handlers[extension]; ok {
		return errors.New("There already is a handler registered for this extension")
	}
	m.handlers[extension] = h
	return nil
}

func (m *wire) Initiate(conn net.Conn) (net.Conn, error) {
	return conn, nil
}

func (m *wire) Handle(conn net.Conn) (net.Conn, error) {
	for {
		line, err := mesh.ReadToken(conn)
		if err != nil {
			// TODO close?
			return nil, err
		}
		m.Process(line)
	}
	return conn, nil
}

func (m *wire) Process(bs []byte) error {
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

func (m *wire) Send(ctx context.Context, extention, payloadType string, payload interface{}, to []string) error {
	if len(to) == 0 {
		return nil
	}

	for _, recipient := range to {
		if recipient == "" {
			continue
		}
		msg := &messageOut{
			Version:     *messagingProtocolVersionCanon,
			Codec:       messagingProtocolCodecJSON,
			Extension:   extention,
			PayloadType: payloadType,
			Payload:     payload,
			From:        m.registry.GetLocalPeerInfo().ID,
			To:          recipient,
		}
		if err := m.sendMessage(ctx, msg); err != nil {
			// TODO Log error
		}
	}

	return nil
}

func (m *wire) Name() string {
	return messagingProtocolName
}

func (m *wire) GetAddresses() []string {
	return []string{m.Name()}
}

func (m *wire) sendMessage(ctx context.Context, msg *messageOut) error {
	logger := m.logger.With(zap.String("peerID", msg.To))
	stream, ok := m.streams[msg.To]
	if !ok || stream == nil {
		conn, err := m.mesh.Dial(ctx, msg.To, messagingProtocolName)
		if err != nil {
			m.logger.Info("could not dial to peer",
				zap.Error(err),
				zap.String("peerID", msg.To),
				zap.Error(err),
			)
			return err
		}

		m.streams[msg.To] = conn
		stream = conn
	}

	logger.Info("Attempting to write message")
	b, err := json.Marshal(msg)
	if err != nil {
		m.logger.Warn("could not marshal outgoing message", zap.Error(err))
		return err
	}

	if err := mesh.WriteToken(stream, b); err != nil {
		m.logger.Warn("could not write outgoing message", zap.Error(err))
		delete(m.streams, msg.To)
		stream.Close()
		return err
	}

	m.logger.Debug("Wrote message", zap.String("peerID", msg.To))
	return nil
}

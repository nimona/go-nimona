package mesh

import (
	"bufio"
	"context"
	"encoding/json"
	"regexp"

	"go.uber.org/zap"

	"github.com/nimona/go-nimona/net"
)

const (
	messagingProtocolName = "messaging"
)

type Messenger interface {
	net.Protocol
	// Publish(msg Message) error
	// Subscribe(topic string) (chan Message, error)
	// Unsubscribe(chan Message) error
}

type messenger struct {
	mesh Mesh

	incomingQueue chan Message
	outgoingQueue chan Message

	subscriptionTopicMatches map[string]*regexp.Regexp
	subscriptions            map[string]map[chan interface{}]bool
	channelSize              int

	logger *zap.Logger
}

func NewMessenger(ms Mesh) (Messenger, error) {
	ctx := context.Background()

	m := &messenger{
		mesh:                     ms,
		incomingQueue:            make(chan Message, 100),
		outgoingQueue:            make(chan Message, 100),
		subscriptionTopicMatches: map[string]*regexp.Regexp{},
		subscriptions:            map[string]map[chan interface{}]bool{},
		channelSize:              100,
		logger:                   net.Logger(ctx).Named("messenger"),
	}

	messages, err := ms.Subscribe("message:.*")
	if err != nil {
		return nil, err
	}

	peerID := ms.GetLocalPeerInfo().ID

	go func() {
		for omsg := range messages {
			msg, ok := omsg.(Message)
			if !ok {
				m.logger.Warn("messenger.ProcessLoop-failed-invalid-message-type")
				continue
			}
			if msg.Nonce == "" {
				msg.Nonce = RandStringBytesMaskImprSrc(8)
			}
			// TODO Hash
			if msg.Sender == peerID {
				m.logger.Info("messenger.ProcessLoop-found-outgoing", zap.String("message", msg.String()))
				m.outgoingQueue <- msg
			} else {
				m.logger.Warn("messenger.ProcessLoop-found-other", zap.String("message", msg.String()))
			}
		}
	}()

	go func() {
		for msg := range m.incomingQueue {
			m.logger.Info("messenger.IncomingLoop", zap.String("message", msg.String()))
			if err := ms.Publish(msg, msg.Topic); err != nil {
				m.logger.Warn("could not publish incoming message", zap.Error(err))
			}
		}
	}()

	go func() {
		for msg := range m.outgoingQueue {
			nctx := context.WithValue(ctx, net.RequestIDKey{}, msg.Nonce)
			logger := m.logger.With(zap.String("req.id", msg.Nonce))
			logger.Info("Attempting to write message", zap.String("peerID", msg.Recipient))
			_, conn, err := m.mesh.Dial(nctx, msg.Recipient, messagingProtocolName)
			if err != nil {
				m.logger.Warn("could not dial to peer", zap.Error(err))
				continue
			}

			b, err := json.Marshal(msg)
			if err != nil {
				m.logger.Warn("could not marshal outgoing message", zap.Error(err))
				continue
			}

			b = append(b, '\n')
			if _, err := conn.Write(b); err != nil {
				m.logger.Warn("could write outgoing message", zap.Error(err))
			}

			m.logger.Debug("Wrote message", zap.String("peerID", msg.Recipient))
		}
	}()

	return m, nil
}

// Negotiate will be called after all the other protocol have been processed
func (m *messenger) Negotiate(fn net.NegotiatorFunc) net.NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c net.Conn) error {
		return fn(ctx, c)
	}
}

// Handle adds the base protocols for transports
func (m *messenger) Handle(fn net.HandlerFunc) net.HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c net.Conn) error {
		scanner := bufio.NewScanner(c)
		for scanner.Scan() {
			line := scanner.Text()
			msg := Message{}
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				return err
			}

			m.incomingQueue <- msg
		}
		return nil
	}
}

func (m *messenger) Name() string {
	return messagingProtocolName
}

func (m *messenger) GetAddresses() []string {
	return []string{m.Name()}
}

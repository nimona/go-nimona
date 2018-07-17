package net

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/ugorji/go/codec"
	"go.uber.org/zap"

	"github.com/nimona/go-nimona/log"
	"github.com/nimona/go-nimona/utils"
)

func init() {
	RegisterContentType("net.handshake", HandshakeMessage{})
}

var (
	// ErrAllAddressesFailed for when a peer cannot be dialed
	ErrAllAddressesFailed = errors.New("all addresses failed to dial")
	// ErrNotForUs message is not meant for us
	ErrNotForUs = errors.New("message not for us")
)

// MessageHandler to handle incoming messages
type MessageHandler func(event *Message) error

// Messenger interface for mocking messenger
type Messenger interface {
	Handle(contentType string, h MessageHandler) error
	Send(ctx context.Context, message *Message, recipients ...string) error
	Listen(ctx context.Context, addrress string) (net.Listener, error)
}

type messenger struct {
	network Networker
	// listener net.Listener

	addressBook PeerManager

	incoming chan net.Conn
	outgoing chan net.Conn
	messages chan *Message
	close    chan bool

	streams  sync.Map
	handlers map[string]MessageHandler
	// handlersLock sync.RWMutex
	logger     *zap.Logger
	streamLock utils.Kmutex
}

// NewMessenger creates a messenger on a given network
func NewMessenger(addressBook *AddressBook) (Messenger, error) {
	ctx := context.Background()

	network, err := NewNetwork(addressBook)
	if err != nil {
		return nil, err
	}

	m := &messenger{
		network:     network,
		addressBook: addressBook,

		incoming: make(chan net.Conn),
		outgoing: make(chan net.Conn),
		messages: make(chan *Message, 100),
		close:    make(chan bool),

		handlers:   map[string]MessageHandler{},
		logger:     log.Logger(ctx).Named("messenger"),
		streamLock: utils.NewKmutex(),
	}

	return m, nil
}

func (w *messenger) Handle(contentType string, h MessageHandler) error {
	if _, ok := w.handlers[contentType]; ok {
		return errors.New("There already is a handler registered for this contentType")
	}
	w.handlers[contentType] = h
	return nil
}

func (w *messenger) Close(peerID string, conn net.Conn) {
	if conn != nil {
		conn.Close()
	}
	w.streams.Range(func(k, v interface{}) bool {
		if k.(string) == peerID {
			w.streams.Delete(k)
		}
		if v.(net.Conn) == conn {
			w.streams.Delete(k)
		}
		return true
	})
}

func (w *messenger) HandleConnection(conn net.Conn) error {
	w.logger.Debug("handling new connection", zap.String("remote", conn.RemoteAddr().String()))

	messageDecoder := codec.NewDecoder(conn, getCborHandler())
	for {
		message := &Message{}
		if err := messageDecoder.Decode(message); err != nil {
			w.logger.Error("could not read message", zap.Error(err))
			w.Close("", conn)
			return err
		}

		w.logger.Debug("handling message", zap.Any("message", message))

		if err := message.Verify(); err != nil {
			w.logger.Warn("could not verify message", zap.Error(err))
			return err
		}

		if message.Type == "net.handshake" {
			payload, ok := message.Payload.(HandshakeMessage)
			if !ok {
				continue
			}
			if err := w.addressBook.PutPeerInfoFromMessage(payload.PeerInfo); err != nil {
				// TODO handle error
				continue
			}
			w.streams.Store(message.Headers.Signer, conn)
			continue
		}

		contentType := message.Type
		var handler MessageHandler
		ok := false
		for handlerContentType, hn := range w.handlers {
			if !strings.HasPrefix(contentType, handlerContentType) {
				continue
			}
			ok = true
			handler = hn
			break
		}

		if !ok {
			w.logger.Info(
				"No handler registered for contentType",
				zap.String("contentType", contentType),
			)
			return nil
		}

		if err := handler(message); err != nil {
			w.logger.Info(
				"Could not handle event",
				zap.String("contentType", contentType),
				zap.Error(err),
			)
			continue
		}
	}
}

// Process incoming message
func (w *messenger) Process(message *Message) error {
	if err := message.Verify(); err != nil {
		w.logger.Warn("could not verify message", zap.Error(err))
		return err
	}

	contentType := message.Type
	var handler MessageHandler
	ok := false
	for handlerContentType, hn := range w.handlers {
		if !strings.HasPrefix(contentType, handlerContentType) {
			continue
		}
		ok = true
		handler = hn
		break
	}

	if !ok {
		w.logger.Info(
			"No handler registered for contentType",
			zap.String("contentType", contentType),
		)
		return nil
	}

	if err := handler(message); err != nil {
		w.logger.Info(
			"Could not handle event",
			zap.String("contentType", contentType),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (w *messenger) Send(ctx context.Context, message *Message, recipients ...string) error {
	signer := w.addressBook.GetLocalPeerInfo()

	if !message.IsSigned() {
		if err := message.Sign(signer); err != nil {
			return err
		}
	}

	if len(recipients) == 0 {
		recipients = message.Headers.Recipients
	}

	for _, recipient := range recipients {
		// TODO deal with error
		if err := w.sendOne(ctx, message, recipient); err != nil {
			// TODO log error
			return err
		}
	}

	return nil
}

func (w *messenger) sendOne(ctx context.Context, message *Message, recipient string) error {
	logger := w.logger.With(zap.String("peerID", recipient))

	w.streamLock.Lock(recipient)
	defer w.streamLock.Unlock(recipient)

	w.logger.Debug("getting conn to write message", zap.String("recipient", recipient))
	conn, err := w.GetOrDial(ctx, recipient)
	if err != nil {
		return err
	}

	// TODO this seems messy
	// try to send the message directly to the recipient
	if err := w.writeMessage(ctx, message, conn); err != nil {
		w.Close(recipient, conn)
		logger.Debug("could not send directly to recipient", zap.Error(err))
	} else {
		return nil
	}

	return ErrAllAddressesFailed
}

func (w *messenger) writeMessage(ctx context.Context, message *Message, rw io.ReadWriter) error {
	signer := w.addressBook.GetLocalPeerInfo()

	if !message.IsSigned() {
		if err := message.Sign(signer); err != nil {
			return err
		}
	}

	messageBytes, err := Marshal(message)
	if err != nil {
		return err
	}

	if _, err := rw.Write(messageBytes); err != nil {
		return err
	}

	w.logger.Debug("writing message", zap.Any("message", message))

	return nil
}

func (w *messenger) GetOrDial(ctx context.Context, peerID string) (net.Conn, error) {
	w.logger.Debug("getting conn", zap.String("peer_id", peerID))
	if peerID == "" {
		return nil, errors.New("missing peer id")
	}

	existingConn, ok := w.streams.Load(peerID)
	if ok {
		return existingConn.(net.Conn), nil
	}

	w.logger.Debug("dialing peer", zap.String("peer_id", peerID))
	conn, err := w.network.Dial(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// TODO move after handshake
	// handle outgoing connections
	w.outgoing <- conn

	// store conn for reuse
	w.streams.Store(peerID, conn)

	w.logger.Debug("writing handshake")

	// handshake so the other side knows who we are
	signer := w.addressBook.GetLocalPeerInfo()
	handshakeMessage := &Message{
		Type: "net.handshake",
		Headers: Headers{
			Recipients: []string{peerID},
		},
		Payload: HandshakeMessage{
			PeerInfo: w.addressBook.GetLocalPeerInfo().Message(),
		},
	}

	// TODO move someplace common with send()
	if err := handshakeMessage.Sign(signer); err != nil {
		return nil, err
	}

	if err := w.writeMessage(ctx, handshakeMessage, conn); err != nil {
		return nil, err
	}

	w.addressBook.PutPeerStatus(peerID, Connected)

	return conn, nil
}

// Listen on an address
// TODO do we need to return a listener?
func (w *messenger) Listen(ctx context.Context, addr string) (net.Listener, error) {
	listener, err := w.network.Listen(ctx, addr)
	if err != nil {
		return nil, err
	}

	closed := false

	go func() {
		for {
			select {
			case message := <-w.messages:
				if err := w.Process(message); err != nil {
					w.logger.Warn("failed to process message", zap.Error(err))
				}
			case conn := <-w.incoming:
				go func() {
					if err := w.HandleConnection(conn); err != nil {
						w.logger.Warn("failed to handle message", zap.Error(err))
					}
				}()
			case conn := <-w.outgoing:
				go func() {
					if err := w.HandleConnection(conn); err != nil {
						w.logger.Warn("failed to handle message", zap.Error(err))
					}
				}()
			case <-w.close:
				closed = true
				w.logger.Debug("connection closed")
				listener.Close()
			}
		}
	}()

	go func() {
		w.logger.Debug("accepting connections", zap.String("address", listener.Addr().String()))
		for {
			conn, err := listener.Accept()
			w.logger.Debug("connection accepted")
			if err != nil {
				if closed {
					return
				}
				w.logger.Error("could not accept", zap.Error(err))
				// TODO check conn is still alive and return
				return
			}
			w.incoming <- conn
		}
	}()

	return listener, nil
}

package net

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	network  Networker
	listener net.Listener

	addressBook PeerManager

	incoming chan net.Conn
	outgoing chan net.Conn
	close    chan bool

	streams      sync.Map
	handlers     map[string]MessageHandler
	handlersLock sync.RWMutex
	logger       *zap.Logger
	streamLock   utils.Kmutex
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
		incoming:    make(chan net.Conn),
		outgoing:    make(chan net.Conn),
		close:       make(chan bool),
		handlers:    map[string]MessageHandler{},
		logger:      log.Logger(ctx).Named("messenger"),
		streamLock:  utils.NewKmutex(),
	}

	return m, nil
}

func (w *messenger) Handle(contentType string, h MessageHandler) error {
	w.handlersLock.Lock()
	defer w.handlersLock.Unlock()
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

func (w *messenger) HandleIncoming(conn net.Conn) error {
	w.logger.Debug("handling new incoming connection", zap.String("remote", conn.RemoteAddr().String()))
	remotePeerID := ""

	messages := make(chan *Message, 100)

	go func() {
		for message := range messages {
			// TODO check recipients
			// if not-ours {
			// 	if err == ErrNotForUs {
			// 		if err := w.ForwardEnvelope(envelope); err != nil {
			// 			w.logger.Error("could not relay envelope", zap.Error(err))
			// 		}
			// 		continue
			// 	}
			// 	w.logger.Error("could not get message from envelope", zap.Error(err))
			// 	continue
			// }
			// TODO make sure this only happens once
			if message.Headers.ContentType == "net.handshake" {
				payload, ok := message.Payload.(*HandshakeMessage)
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
			if err := w.Process(message); err != nil {
				// w.logger.Error("could not process line", zap.Error(err))
				// TODO should we return err?
				w.logger.Error("could not process message", zap.Error(err))
				w.Close(remotePeerID, conn)
				// return err
			}
		}
	}()

	defer close(messages)

	messageDecoder := codec.NewDecoder(conn, &codec.CborHandle{})

	defer w.Close(remotePeerID, conn)

	for {
		// timeoutDuration := time.Minute
		// conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		message := &Message{}
		if err := messageDecoder.Decode(message); err != nil {
			// if err != io.EOF {
			w.logger.Error("could not read message", zap.Error(err))
			// }
			fmt.Println("3")
			return err
			// continue
		}

		fmt.Println("DECC")

		mm, err := Marshal(message)
		if err != nil {
			fmt.Println("4")
			return err
		}

		mo, err := Unmarshal(mm)
		if err != nil {
			fmt.Println("5")
			return err
		}

		// TODO fix verification
		// b, err := json.MarshalIndent(mo, "", "  ")
		// fmt.Println("MSG", string(b), err)
		// if err := message.Verify(); err != nil {
		// 	w.logger.Warn("could not verify message", zap.Error(err))
		// 	continue
		// }
		messages <- mo
	}
}

func (w *messenger) HandleOutgoing(conn net.Conn) error {
	w.logger.Debug("handling new outgoing connection", zap.String("remote", conn.RemoteAddr().String()))

	messages := make(chan *Message, 100)

	go func() {
		for message := range messages {
			// TODO check if message is not for us
			// if err != nil {
			// 	if err == ErrNotForUs {
			// 		if err := w.ForwardEnvelope(envelope); err != nil {
			// 			w.logger.Error("could not relay envelope", zap.Error(err))
			// 		}
			// 		continue
			// 	}
			// 	w.logger.Error("could not get message from envelope", zap.Error(err))
			// 	continue
			// }
			if err := w.Process(message); err != nil {
				// TODO should we return err?
				w.logger.Error("could not process message", zap.Error(err))
				w.Close("", conn)
				// return err
			}
		}
	}()

	defer close(messages)

	msgpackHandler := new(codec.MsgpackHandle)
	msgpackHandler.RawToString = false

	messageDecoder := codec.NewDecoder(conn, &codec.CborHandle{})

	for {
		message := &Message{}
		if err := messageDecoder.Decode(message); err != nil {
			w.logger.Error("could not read message", zap.Error(err))
			w.Close("", conn)
			return err
		}
		messages <- message
	}
}

// func (w *messenger) ForwardMessage(message *Message) error {
// 	messageDecoder := codec.NewDecoder(conn, &codec.CborHandle{})

// 	// decrypt message
// 	r := bytes.NewReader(envelope.Message)
// 	mki, _, err := saltpack.NewDecryptStream(saltpack.SingleVersionValidator(saltpack.CurrentVersion()), r, w.keyring)
// 	if err != nil {
// 		if err == saltpack.ErrNoDecryptionKey {
// 			if len(mki.NamedReceivers) == 0 {
// 				return errors.New("no recipients")
// 			}
// 			for _, receiver := range mki.NamedReceivers {
// 				recipientID := fmt.Sprintf("%x", receiver)
// 				ctx := context.Background() // TODO Fix context
// 				encodedEnvelope := []byte{}
// 				envelopeEncoder := codec.NewEncoderBytes(&encodedEnvelope, msgpackHandler)
// 				if err := envelopeEncoder.Encode(envelope); err != nil {
// 					return err
// 				}
// 				if err := w.writeMessage(ctx, encodedEnvelope, recipientID); err != nil {
// 					return err
// 				}
// 			}
// 			return nil
// 		}
// 		return err
// 	}

// 	return errors.New("message is meant for us, no need to forward")
// }

// Process incoming message
// TODO Process from channel, don't call directly so we don't have to lock handlers
func (w *messenger) Process(message *Message) error {
	w.handlersLock.RLock()
	defer w.handlersLock.RUnlock()
	contentType := message.Headers.ContentType
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
		return errors.New("could not handle")
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
	logger.Debug("sending message", zap.String("contentType", message.Headers.ContentType))

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

	b, _ := json.MarshalIndent(message, "", "  ")
	fmt.Println("message", string(b))

	messageBytes, err := Marshal(message)
	if err != nil {
		return err
	}

	if _, err := rw.Write(messageBytes); err != nil {
		return err
	}

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
		Headers: Headers{
			ContentType: "net.handshake",
			Recipients:  []string{peerID},
		},
		Payload: &HandshakeMessage{
			PeerInfo: w.addressBook.GetLocalPeerInfo().Message(),
		},
	}

	// TODO move someplace common with send()
	if err := handshakeMessage.Sign(signer); err != nil {
		return nil, err
	}

	b, _ := json.MarshalIndent(handshakeMessage, "", "  ")
	fmt.Println("HandshakeMessage", string(b))

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
			case conn := <-w.incoming:
				go func() {
					err := w.HandleIncoming(conn)
					fmt.Println("failed to HandleIncoming()", err)
				}()
			case conn := <-w.outgoing:
				go func() {
					err := w.HandleOutgoing(conn)
					fmt.Println("failed to HandleOutgoing()", err)
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

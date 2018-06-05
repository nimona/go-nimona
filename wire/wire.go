package wire

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	igd "github.com/emersion/go-upnp-igd"
	"github.com/keybase/saltpack"
	"github.com/keybase/saltpack/basic"
	"github.com/ugorji/go/codec"
	"go.uber.org/zap"

	"github.com/nimona/go-nimona/log"
	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/utils"
)

var (
	// ErrAllAddressesFailed for when a peer cannot be dialed
	ErrAllAddressesFailed = errors.New("all addresses failed to dial")
	// ErrNotForUs message is not meant for us
	ErrNotForUs = errors.New("message not for us")
)

// EventHandler for wire.HandleExtensionEvents
type EventHandler func(event *Message) error

// Wire interface for mocking wire
type Wire interface {
	HandleExtensionEvents(extension string, h EventHandler) error
	Send(ctx context.Context, extension, payloadType string,
		payload interface{}, recipients []string) error
	Pack(extension, payloadType string, payload interface{},
		recipient string, hideSender, hideRecipients bool) ([]byte, error)
	Listen(addr string) (net.Listener, string, error)
}

type wire struct {
	incoming   chan net.Conn
	outgoing   chan net.Conn
	close      chan bool
	keyring    *basic.Keyring
	registry   mesh.Registry
	streams    sync.Map
	handlers   map[string]EventHandler
	logger     *zap.Logger
	streamLock utils.Kmutex
}

// NewWire creates a new wire protocol based on a registry
func NewWire(reg mesh.Registry) (Wire, error) {
	ctx := context.Background()

	w := &wire{
		incoming: make(chan net.Conn),
		outgoing: make(chan net.Conn),
		close:    make(chan bool),
		keyring:  reg.GetKeyring(),
		registry: reg,
		// streams:  map[string]io.ReadWriteCloser{},
		handlers:   map[string]EventHandler{},
		logger:     log.Logger(ctx).Named("wire"),
		streamLock: utils.NewKmutex(),
	}

	return w, nil
}

func (w *wire) HandleExtensionEvents(extension string, h EventHandler) error {
	if _, ok := w.handlers[extension]; ok {
		return errors.New("There already is a handler registered for this extension")
	}
	w.handlers[extension] = h
	return nil
}

func (w *wire) Close(peerID string, conn net.Conn) {
	// ctx := context.Background()

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

	// if peerID != "" {
	// 	go w.GetOrDial(ctx, peerID)
	// }
}

func (w *wire) HandleIncoming(conn net.Conn) error {
	w.logger.Debug("handling new incoming connection")
	handshakeComplete := false
	remotePeerID := ""

	for {
		timeoutDuration := time.Minute
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		message, err := ReadMessage(conn)
		if err != nil {
			if err != io.EOF {
				w.logger.Error("could not read line", zap.Error(err))
				w.Close(remotePeerID, conn)
			}
			return err
		}
		if !handshakeComplete {
			receivedPeerID, err := w.ProcessHandshake(message)
			if err != nil {
				w.logger.Error("could not handle handlshake", zap.Error(err))
				w.Close(remotePeerID, conn)
				return err
			}
			remotePeerID = receivedPeerID
			w.streams.Store(remotePeerID, conn)
			handshakeComplete = true
			continue
		}
		if err := w.Process(message); err != nil {
			// w.logger.Error("could not process line", zap.Error(err))
			// TODO should we return err?
			w.Close(remotePeerID, conn)
			return err
		}
	}
}

func (w *wire) HandleOutgoing(conn net.Conn) error {
	w.logger.Debug("handling new outgoing connection")

	for {
		timeoutDuration := time.Minute
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		message, err := ReadMessage(conn)
		if err != nil {
			if err != io.EOF {
				w.logger.Error("could not read line", zap.Error(err))
				w.Close("", conn)
			}
			return err
		}
		if err := w.Process(message); err != nil {
			// w.logger.Error("could not process line", zap.Error(err))
			// TODO should we return err?
			w.Close("", conn)
			return err
		}
	}
}

func (w *wire) ProcessHandshake(encryptedBody []byte) (string, error) {
	msg, err := w.DecodeMessage(encryptedBody)
	if err != nil {
		return "", err
	}

	handshake := &handshakeMessage{}
	if err := msg.DecodePayload(&handshake); err != nil {
		return "", err
	}

	return handshake.PeerID, nil
}

func (w *wire) Process(encryptedBody []byte) error {
	msg, err := w.DecodeMessage(encryptedBody)
	if err != nil {
		if err == ErrNotForUs {
			return nil
		}
		return err
	}

	hn, ok := w.handlers[msg.Extension]
	if !ok {
		w.logger.Info(
			"No handler registered for extension",
			zap.String("extension", msg.Extension),
		)
		return errors.New("no handler")
	}

	if err := hn(msg); err != nil {
		w.logger.Info(
			"Could not handle event",
			zap.String("extension", msg.Extension),
			zap.String("payload_type", msg.PayloadType),
			zap.Error(err),
		)
		return errors.New("could not handle")
	}

	return nil
}

func (w *wire) DecodeMessage(encryptedBody []byte) (*Message, error) {
	// mki, body, _, err := saltpack.Dearmor62DecryptOpen(saltpack.CheckKnownMajorVersion, encryptedBody, w.keyring)
	r := bytes.NewReader(encryptedBody)
	mki, pr, err := saltpack.NewDecryptStream(saltpack.SingleVersionValidator(saltpack.CurrentVersion()), r, w.keyring)
	if err != nil {
		if err == saltpack.ErrNoDecryptionKey {
			if len(mki.NamedReceivers) > 0 {
				for _, receiver := range mki.NamedReceivers {
					recipientID := fmt.Sprintf("%x", receiver)
					ctx := context.Background() // TODO Fix context
					if fwErr := w.sendMessage(ctx, []byte(encryptedBody), recipientID); fwErr != nil {
						w.logger.Error("could not relay message", zap.Error(fwErr))
					}
				}
			}
			return nil, ErrNotForUs
		}
		return nil, err
	}

	body, errRead := ioutil.ReadAll(pr)
	if errRead != nil {
		return nil, errRead
	}

	msg := &Message{}
	h := new(codec.MsgpackHandle)
	h.RawToString = false
	enc := codec.NewDecoderBytes(body, h)
	if err := enc.Decode(&msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (w *wire) Send(ctx context.Context, extension, payloadType string, payload interface{}, recipients []string) error {
	for _, recipient := range recipients {
		// TODO deal with error
		if err := w.SendOne(ctx, extension, payloadType, payload, recipient); err != nil {
			return err
		}
	}
	return nil
}

func (w *wire) SendOne(ctx context.Context, extension, payloadType string, payload interface{}, recipient string) error {
	logger := w.logger.With(zap.String("peerID", recipient))
	logger.Debug("sending message", zap.String("extention", extension), zap.String("payloadType", payloadType))

	// create message for recipient
	msg, err := w.Pack(extension, payloadType, payload, recipient, true, true)
	if err != nil {
		return err
	}

	// try to send the message directly to the recipient
	if err := w.sendMessage(ctx, msg, recipient); err != nil {
		logger.Debug("could not send directly to recipient", zap.Error(err))
	} else {
		return nil
	}

	// else, try to find a relay peer and send it to them
	pi, err := w.registry.GetPeerInfo(recipient)
	if err != nil {
		logger.Debug("could not get relay for recipient", zap.Error(err))
		return err
	}

	// go through the addresses and find any relays
	relayIDs := []string{}
	for _, address := range pi.Addresses {
		if strings.HasPrefix(address, "relay:") {
			relayIDs = append(relayIDs, strings.Replace(address, "relay:", "", 1))
		}
	}

	// if no relays found, fail
	if len(relayIDs) == 0 {
		return ErrAllAddressesFailed
	}

	// so, we need to create message for recipient, but not anonymous
	msg, err = w.Pack(extension, payloadType, payload, recipient, false, false)
	if err != nil {
		logger.Debug("could not pack message for relay", zap.Error(err))
		return err
	}

	// go through the relays and send the message
	for _, relayID := range relayIDs {
		if err := w.sendMessage(ctx, msg, relayID); err != nil {
			logger.Debug("could not send to relay", zap.Error(err))
		} else {
			return nil
		}
	}

	// else fail
	return ErrAllAddressesFailed
}

// Pack a message into something we can send
func (w *wire) Pack(extension, payloadType string, payload interface{},
	recipient string, hideSender, hideRecipients bool) ([]byte, error) {
	hideRecipients = false
	pks := []saltpack.BoxPublicKey{}
	pi, err := w.registry.GetPeerInfo(recipient)
	if err != nil {
		return nil, err
	}
	if hideRecipients {
		pks = append(pks, &HiddenPublicKey{pi.GetPublicKey()})
	} else {
		pks = append(pks, pi.GetPublicKey())
	}

	out := &Message{
		Codec:       "json",
		Extension:   extension,
		PayloadType: payloadType,
	}
	if err := out.EncodePayload(payload); err != nil {
		return nil, err
	}

	encodedMsg := []byte{}
	h := new(codec.MsgpackHandle)
	h.RawToString = false
	enc := codec.NewEncoderBytes(&encodedMsg, h)
	if err := enc.Encode(out); err != nil {
		return nil, err
	}

	var sk saltpack.BoxSecretKey
	sk = w.registry.GetLocalPeerInfo().GetSecretKey()
	if hideSender {
		sk = &HiddenSecretKey{sk}
	}

	xpi := w.registry.GetLocalPeerInfo()

	if bb, ok := payload.([]byte); ok {
		var ciphertext bytes.Buffer
		rw, err := saltpack.NewEncryptStream(saltpack.CurrentVersion(), &ciphertext, xpi.GetSecretKey(), pks)
		if err != nil {
			return nil, err
		}

		if _, err := rw.Write(bb); err != nil {
			return nil, err
		}

		if err := rw.Close(); err != nil {
			return nil, err
		}
		fmt.Printf("%d, %d\n", len(bb), len(ciphertext.Bytes()))
	}
	// sender := xpi.GetSecretKey()

	var ciphertext bytes.Buffer
	rw, err := saltpack.NewEncryptStream(saltpack.CurrentVersion(), &ciphertext, xpi.GetSecretKey(), pks)
	if err != nil {
		return nil, err
	}

	if _, err := rw.Write(encodedMsg); err != nil {
		return nil, err
	}

	if err := rw.Close(); err != nil {
		return nil, err
	}

	// plainsink, err := saltpack.NewSigncryptSealStream(&ciphertext, w.keyring, sender, pks, arg.SymmetricReceivers)
	// ecryptedMsg, err := saltpack.EncryptArmor62Seal(saltpack.CurrentVersion(), encodedMsg, sk, pks, "WIRE")
	// if err != nil {
	// 	return nil, err
	// }
	// return []byte(ecryptedMsg), nil

	return ciphertext.Bytes(), nil
}

func (w *wire) sendMessage(ctx context.Context, msg []byte, peerID string) error {
	conn, err := w.GetOrDial(ctx, peerID)
	if err != nil {
		return err
	}

	if err := WriteMessage(conn, []byte(msg)); err != nil {
		w.Close(peerID, conn)
		return err
	}

	return nil
}

func (w *wire) GetOrDial(ctx context.Context, peerID string) (net.Conn, error) {
	if peerID == "" {
		return nil, errors.New("missing peer id")
	}

	existingConn, ok := w.streams.Load(peerID)
	if ok {
		return existingConn.(net.Conn), nil
	}

	w.streamLock.Lock(peerID)
	defer w.streamLock.Unlock(peerID)

	newConn, err := w.Dial(ctx, peerID)
	if err != nil {
		return nil, err
	}

	w.streams.Store(peerID, newConn)
	return newConn, nil
}

// Dial to a peer and return a net.Conn
func (w *wire) Dial(ctx context.Context, peerID string) (net.Conn, error) {
	peerInfo, err := w.registry.GetPeerInfo(peerID)
	if err != nil {
		return nil, err
	}

	var conn net.Conn
	for _, addr := range peerInfo.Addresses {
		if !strings.HasPrefix(addr, "tcp:") {
			continue
		}
		addr = strings.Replace(addr, "tcp:", "", 1)
		dialer := net.Dialer{Timeout: time.Second * 5}
		newConn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			// TODO(superdecimal) Address can be black-listed maybe?
			continue
		}
		conn = newConn
		break
	}
	if conn == nil {
		// TODO(superdecimal) Mark peer as non-connectable directly
		return nil, ErrAllAddressesFailed
	}

	// handle outgoing connections
	w.outgoing <- conn

	// handshake so the other side knows who we are
	w.streams.Store(peerID, conn)
	handshake := &handshakeMessage{
		PeerID: w.registry.GetLocalPeerInfo().ID,
	}
	if err := w.SendOne(ctx, "wire", "handshake", handshake, peerID); err != nil {
		// TODO close and remove stream
		return nil, err
	}

	// TODO(superdecimal) Mark peer as connectable directly

	return conn, nil
}

// Listen on an address
// TODO do we need to return a listener?
func (w *wire) Listen(addr string) (net.Listener, string, error) {
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", err
	}
	port := tcpListener.Addr().(*net.TCPAddr).Port

	lock := sync.Mutex{}
	addresses := GetAddresses(tcpListener)

	devices := make(chan igd.Device)
	go func() {
		for device := range devices {
			upnp := true
			upnpFlag := os.Getenv("UPNP")
			if upnpFlag != "" {
				upnp, _ = strconv.ParseBool(upnpFlag)
			}
			if !upnp {
				continue
			}
			externalAddress, err := device.GetExternalIPAddress()
			if err != nil {
				w.logger.Error("could not get external ip", zap.Error(err))
				continue
			}
			desc := "nimona"
			ttl := time.Hour * 24 * 365
			if _, err := device.AddPortMapping(igd.TCP, port, port, desc, ttl); err != nil {
				w.logger.Error("could not add port mapping", zap.Error(err))
			} else {
				lock.Lock()
				addresses = append(addresses, fmt.Sprintf("tcp:%s:%d", externalAddress.String(), port))
				lock.Unlock()
			}
		}
	}()

	if err := igd.Discover(devices, 5*time.Second); err != nil {
		w.logger.Error("could not discover devices", zap.Error(err))
	}

	addresses = append(addresses, "relay:7730b73e34ae2e3ad92235aefc7ee0366736602f96785e6f35e8b710923b4562")
	w.registry.GetLocalPeerInfo().UpdateAddresses(addresses)

	go func() {
		for {
			conn := <-w.incoming
			go w.HandleIncoming(conn)
		}
	}()

	go func() {
		for {
			conn := <-w.outgoing
			go w.HandleOutgoing(conn)
		}
	}()

	closed := false

	go func() {
		closed = true
		<-w.close
		// TODO do we need to do something here?
		w.logger.Debug("connection closed")
		tcpListener.Close()
	}()

	go func() {
		for {
			conn, err := tcpListener.Accept()
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

	return nil, tcpListener.Addr().String(), nil
}

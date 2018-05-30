package wire

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	igd "github.com/emersion/go-upnp-igd"
	"github.com/keybase/saltpack"
	"github.com/keybase/saltpack/basic"
	"go.uber.org/zap"

	"github.com/nimona/go-nimona/log"
	"github.com/nimona/go-nimona/mesh"
)

var (
	// ErrAllAddressesFailed for when a peer cannot be dialed
	ErrAllAddressesFailed = errors.New("all addresses failed to dial")
)

// EventHandler for wire.HandleExtensionEvents
type EventHandler func(event *Message) error

// Wire interface for mocking wire
type Wire interface {
	HandleExtensionEvents(extension string, h EventHandler) error
	Send(ctx context.Context, extension, payloadType string,
		payload interface{}, recipients []string) error
	Pack(saltpacked bool, extension, payloadType string, payload interface{},
		recipient string, hideSender, hideRecipients bool) (string, error)
	Listen(addr string) (net.Listener, string, error)
}

type wire struct {
	accepted chan net.Conn
	close    chan bool
	keyring  *basic.Keyring
	registry mesh.Registry
	streams  map[string]io.ReadWriteCloser
	handlers map[string]EventHandler
	logger   *zap.Logger
}

// NewWire creates a new wire protocol based on a registry
func NewWire(reg mesh.Registry) (Wire, error) {
	ctx := context.Background()

	w := &wire{
		accepted: make(chan net.Conn),
		close:    make(chan bool),
		keyring:  mesh.Keyring,
		registry: reg,
		streams:  map[string]io.ReadWriteCloser{},
		handlers: map[string]EventHandler{},
		logger:   log.Logger(ctx).Named("wire"),
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

func (w *wire) Handle(conn net.Conn) error {
	w.logger.Debug("handling new connection")
	handshakeComplete := false
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if !handshakeComplete {
			remotePeerID, err := w.ProcessHandshake(line)
			if err != nil {
				w.logger.Error("could not handle handlshake", zap.Error(err))
				return err
			}
			// TODO Lock
			w.streams[remotePeerID] = conn
			handshakeComplete = true
			continue
		}
		if err := w.Process(line); err != nil {
			w.logger.Error("could not process line", zap.Error(err), zap.String("line", line))
			// TODO should we return err?
		}
	}

	return nil
}

func (w *wire) ProcessHandshake(encryptedBody string) (string, error) {
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

func (w *wire) Process(encryptedBody string) error {
	msg, err := w.DecodeMessage(encryptedBody)
	if err != nil {
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

func (w *wire) DecodeMessage(encryptedBody string) (*Message, error) {
	mki, body, _, err := saltpack.Dearmor62DecryptOpen(saltpack.CheckKnownMajorVersion, encryptedBody, w.keyring)
	if err != nil {
		if err == saltpack.ErrNoDecryptionKey {
			if len(mki.NamedReceivers) > 0 {
				for _, receiver := range mki.NamedReceivers {
					recipientID := fmt.Sprintf("%x", receiver)
					ctx := context.Background() // TODO Fix context
					if fwErr := w.sendMessage(ctx, encryptedBody+"\n", recipientID); fwErr != nil {
						w.logger.Error("could not relay message", zap.Error(fwErr))
						return nil, fwErr // TODO Should we return an error eventually?
					}
				}
			}
			return nil, errors.New("Message was not for us")
		}
		return nil, err
	}

	msg := &Message{}
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		w.logger.Info(
			"Could not unmarshal  into type",
			zap.String("extension", msg.Extension),
			zap.String("payload_type", msg.PayloadType),
			zap.Error(err),
		)
		return nil, errors.New("could not unmarshal into type")
	}

	return msg, nil
}

func (w *wire) Send(ctx context.Context, extension, payloadType string, payload interface{}, recipients []string) error {
	for _, recipient := range recipients {
		// TODO deal with error
		w.SendOne(ctx, extension, payloadType, payload, recipient)
	}
	return nil
}

func (w *wire) SendOne(ctx context.Context, extension, payloadType string, payload interface{}, recipient string) error {
	saltpacked := true

	// create message for recipient
	msg, err := w.Pack(saltpacked, extension, payloadType, payload, recipient, true, true)
	if err != nil {
		return err
	}

	// try to send the message directly to the recipient
	if err := w.sendMessage(ctx, msg, recipient); err != nil {
		w.logger.Debug("could not send directly to recipient", zap.String("recipient", recipient))
	} else {
		return nil
	}

	// else, try to find a relay peer and send it to them
	relayIDs := []string{}
	pi, err := w.registry.GetPeerInfo(recipient)
	if err != nil {
		w.logger.Debug("could not get relay for recipient", zap.String("recipient", recipient))
		return err
	}

	// go through the addresses and find any relays
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
	msg, err = w.Pack(saltpacked, extension, payloadType, payload, recipient, false, false)
	if err != nil {
		w.logger.Debug("could not pack message for relay", zap.Error(err))
		return err
	}

	// go through the relays and send the message
	for _, relayID := range relayIDs {
		if err := w.sendMessage(ctx, msg, relayID); err != nil {
			w.logger.Debug("could not send to relay",
				zap.String("recipient", recipient), zap.Error(err))
		} else {
			return nil
		}
	}

	// else fail
	return ErrAllAddressesFailed
}

func (w *wire) Pack(saltpacked bool, extension, payloadType string,
	payload interface{}, recipient string, hideSender, hideRecipients bool) (string, error) {
	pks := []saltpack.BoxPublicKey{}
	pi, err := w.registry.GetPeerInfo(recipient)
	if err != nil {
		return "", err
	}
	if hideRecipients {
		pks = append(pks, &HiddenPublicKey{pi.GetPublicKey()})
	} else {
		pks = append(pks, pi.GetPublicKey())
	}

	var out interface{}

	msg := &messageOut{
		Extension:   extension,
		PayloadType: payloadType,
		Payload:     payload,
	}

	if !saltpacked {
		out = map[string]interface{}{
			"to":      recipient,
			"from":    w.registry.GetLocalPeerInfo().ID,
			"payload": msg,
		}
	} else {
		out = msg
	}

	encodedMsg, err := json.Marshal(out)
	if err != nil {
		return "", err
	}

	if !saltpacked {
		return string(encodedMsg), nil
	}

	var sk saltpack.BoxSecretKey
	sk = w.registry.GetLocalPeerInfo().GetSecretKey()
	if hideSender {
		sk = &HiddenSecretKey{sk}
	}
	ecryptedMsg, err := saltpack.EncryptArmor62Seal(saltpack.CurrentVersion(), encodedMsg, sk, pks, "WIRE")
	if err != nil {
		return "", err
	}

	return ecryptedMsg, nil
}

func (w *wire) sendMessage(ctx context.Context, msg string, recipient string) error {
	logger := w.logger.With(zap.String("peerID", recipient))

	conn, ok := w.streams[recipient]
	if !ok || conn == nil {
		var err error
		conn, err = w.Dial(ctx, recipient)
		if err != nil {
			logger.Info("could not dial to peer", zap.Error(err))
			return err
		}
	}

	if _, err := conn.Write([]byte(msg)); err != nil {
		// TODO Lock
		logger.Warn("could not write outgoing message", zap.Error(err))
		delete(w.streams, recipient)
		conn.Close()
		return err
	}

	logger.Debug("Wrote message")
	return nil
}

func (w *wire) Dial(ctx context.Context, peerID string) (net.Conn, error) {
	logger := w.logger.With(zap.String("peer_id", peerID))
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
			logger.Debug("could not dial", zap.Error(err), zap.String("address", addr))
			continue
		}
		conn = newConn
		break
	}
	if conn == nil {
		// TODO(superdecimal) Mark peer as non-connectable directly
		return nil, ErrAllAddressesFailed
	}

	// handle incoming connections
	w.accepted <- conn

	// handshake so the other side knows who we are
	w.streams[peerID] = conn
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
			conn := <-w.accepted
			go w.Handle(conn)
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
			w.accepted <- conn
		}
	}()

	return nil, tcpListener.Addr().String(), nil
}

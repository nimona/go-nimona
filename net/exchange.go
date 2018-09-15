package net

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	ucodec "github.com/ugorji/go/codec"
	"go.uber.org/zap"

	"nimona.io/go/base58"
	"nimona.io/go/blocks"
	"nimona.io/go/codec"
	"nimona.io/go/crypto"
	"nimona.io/go/log"
	"nimona.io/go/peers"
	"nimona.io/go/storage"
	"nimona.io/go/utils"
)

var (
	// ErrInvalidRequest when received an invalid request block
	ErrInvalidRequest = errors.New("Invalid request")
)

var (
	// ErrAllAddressesFailed for when a peer cannot be dialed
	ErrAllAddressesFailed = errors.New("all addresses failed to dial")
	// ErrNoAddresses for when a peer has no addresses
	ErrNoAddresses = errors.New("no addresses")
	// ErrNotForUs block is not meant for us
	ErrNotForUs = errors.New("block not for us")
)

// Exchange interface for mocking exchange
type Exchange interface {
	Get(ctx context.Context, id string) (interface{}, error)
	Handle(contentType string, h func(o blocks.Typed) error) error
	Send(ctx context.Context, o blocks.Typed, recipient *crypto.Key, opts ...blocks.PackOption) error
	Listen(ctx context.Context, addrress string) (net.Listener, error)
	RegisterDiscoverer(discovery Discoverer)
}

type exchange struct {
	network     Networker
	addressBook *peers.AddressBook
	discovery   Discoverer

	outgoingPayloads chan outBlock
	incoming         chan net.Conn
	outgoing         chan net.Conn
	close            chan bool

	streams    sync.Map
	handlers   []handler
	logger     *zap.Logger
	streamLock utils.Kmutex

	store         storage.Storage
	getRequests   sync.Map
	subscriptions sync.Map
}

type outBlock struct {
	recipient *crypto.Key
	bytes     []byte
}

type incBlock struct {
	peerID  string
	conn    net.Conn
	payload []byte
}

type handler struct {
	contentType glob.Glob
	handler     func(o blocks.Typed) error
}

// NewExchange creates a exchange on a given network
func NewExchange(addressBook *peers.AddressBook, store storage.Storage) (Exchange, error) {
	ctx := context.Background()

	network, err := NewNetwork(addressBook)
	if err != nil {
		return nil, err
	}

	w := &exchange{
		network:     network,
		addressBook: addressBook,

		outgoingPayloads: make(chan outBlock, 100),
		incoming:         make(chan net.Conn),
		outgoing:         make(chan net.Conn),
		close:            make(chan bool),

		handlers:   []handler{},
		logger:     log.Logger(ctx).Named("exchange"),
		streamLock: utils.NewKmutex(),

		store:       store,
		getRequests: sync.Map{},
	}

	self := w.addressBook.GetLocalPeerInfo()
	key := w.addressBook.GetLocalPeerKey()

	go func() {
		for block := range w.outgoingPayloads {
			recipientThumbPrint := block.recipient.Thumbprint()
			if self.Thumbprint() == recipientThumbPrint {
				w.logger.Info("cannot send block to self")
				continue
			}

			// TODO log error and reconsider the async
			// TODO also maybe we need to verify it or something?

			logger := w.logger.With(zap.String("peerID", recipientThumbPrint))

			// try to send the block directly to the recipient
			logger.Debug("getting conn to write block")
			conn, err := w.GetOrDial(ctx, recipientThumbPrint)
			if err != nil {
				logger.Debug("could not get conn to recipient", zap.Error(err))
			} else {
				if err := w.writeBlock(ctx, block.bytes, conn); err != nil {
					// TODO better handling of connection errors
					w.Close(recipientThumbPrint, conn)
					logger.Debug("could not write to recipient", zap.Error(err))
				} else {
					// update peer status
					w.addressBook.PutPeerStatus(recipientThumbPrint, peers.StatusConnected)
					continue
				}
			}

			// else try to send message via their relay addresses
			conn, err = w.getOrDialRelay(ctx, recipientThumbPrint)
			if err != nil {
				logger.Debug("could not get conn to recipient's relay", zap.Error(err))
				continue
			}

			// create forwarded block
			fw := &ForwardRequest{
				Recipient: block.recipient,
				Block:     block.bytes,
			}

			fwb, err := blocks.PackEncode(fw, blocks.SignWith(key))
			if err != nil {
				panic(err)
				continue
			}

			// try to send the block directly to the recipient
			if err := w.writeBlock(ctx, fwb, conn); err != nil {
				// TODO better handling of connection errors
				// TODO this is a bad close, id is of recipient, conn is of relay
				w.Close(recipientThumbPrint, conn)
				logger.Debug("could not write to relay", zap.Error(err))
				// update peer status
				w.addressBook.PutPeerStatus(recipientThumbPrint, peers.StatusError)
				continue
			}

			// update peer status
			w.addressBook.PutPeerStatus(recipientThumbPrint, peers.StatusCanConnect)
		}
	}()

	return w, nil
}

func (w *exchange) RegisterDiscoverer(discovery Discoverer) {
	w.discovery = discovery

	ctx := context.Background()
	go func() {
		for {
			blocks, err := w.store.List()
			if err != nil {
				time.Sleep(time.Second * 10)
				continue
			}

			for _, block := range blocks {
				if err := w.discovery.PutProviders(ctx, block); err != nil {
					w.logger.Warn("could not announce provider for block", zap.String("id", block))
				}
			}

			time.Sleep(time.Second * 30)
		}
	}()
}

func (w *exchange) Handle(typePatern string, h func(o blocks.Typed) error) error {
	g, err := glob.Compile(typePatern, '.', '/', '#')
	if err != nil {
		return err
	}
	w.handlers = append(w.handlers, handler{
		contentType: g,
		handler:     h,
	})
	return nil
}

func (w *exchange) Close(peerID string, conn net.Conn) {
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

func (w *exchange) HandleConnection(conn net.Conn) error {
	w.logger.Debug("handling new connection", zap.String("remote", conn.RemoteAddr().String()))

	pDecoder := ucodec.NewDecoder(conn, codec.CborHandler())
	for {
		p := &blocks.Block{}
		if err := pDecoder.Decode(&p); err != nil {
			w.logger.Error("could not read block", zap.Error(err))
			w.Close("", conn)
			return err
		}

		if err := w.process(p, conn); err != nil {
			w.Close("", conn)
			return err
		}
	}
}

// Process incoming block
func (w *exchange) process(block *blocks.Block, conn net.Conn) error {
	blockBytes, err := blocks.Encode(block)
	if err != nil {
		return err
	}

	payload, err := blocks.Unpack(block, blocks.Verify())
	if err != nil {
		return err
	}

	if os.Getenv("DEBUG_BLOCKS") != "" {
		fmt.Println("Processing type", block.Type)
		fmt.Println("< ---------- inc block / start")
		b, _ := json.MarshalIndent(block, "< ", "  ")
		fmt.Println(string(b))
		fmt.Println("< ---------- inc block / end")
	}

	SendBlockEvent(
		"incoming",
		block.Type,
		len(blockBytes),
	)

	blockID := block.ID()
	if blocks.ShouldPersist(block.Type) {
		if err := w.store.Store(blockID, blockBytes); err != nil {
			if err != storage.ErrExists {
				w.logger.Warn("could not write block", zap.Error(err))
			}
		}
	}

	// TODO convert these into proper handlers
	contentType := block.Type
	switch payload := payload.(type) {
	case *ForwardRequest:
		w.logger.Info("got forwarded message", zap.String("recipient", payload.Recipient.Thumbprint()))
		w.outgoingPayloads <- outBlock{
			recipient: payload.Recipient,
			bytes:     payload.Block,
		}
		return nil

	case *BlockRequest:
		if err := w.handleRequestBlock(payload); err != nil {
			w.logger.Warn("could not handle request block", zap.Error(err))
		}

	case *BlockResponse:
		if err := w.handleBlockResponse(payload); err != nil {
			w.logger.Warn("could not handle request block", zap.Error(err))
		}

	case *HandshakePayload:
		if err := w.addressBook.PutPeerInfo(payload.PeerInfo); err != nil {
			return err
		}

		w.streams.Store(payload.PeerInfo.Thumbprint(), conn)
		return nil
	}

	for _, handler := range w.handlers {
		if handler.contentType.Match(contentType) {
			if err := handler.handler(payload); err != nil {
				w.logger.Info(
					"Could not handle event",
					zap.String("contentType", contentType),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

func (w *exchange) handleBlockResponse(payload *BlockResponse) error {
	// Check if nonce exists in local addressBook
	value, ok := w.getRequests.Load(payload.RequestID)
	if !ok {
		return nil
	}

	req, ok := value.(*BlockRequest)
	if !ok {
		return ErrInvalidRequest
	}

	block, err := blocks.Decode(payload.Block)
	if err != nil {
		panic(err)
		return err
	}

	if blocks.ShouldPersist(block.Type) {
		blockID, _ := blocks.SumSha3(payload.Block)
		w.store.Store(blockID, payload.Block)
	}

	p, err := blocks.Unpack(block)
	if err != nil {
		return err
	}

	req.response <- p
	return nil
}

func (w *exchange) handleRequestBlock(payload *BlockRequest) error {
	blockBytes, err := w.store.Get(payload.ID)
	if err != nil {
		return err
	}

	// TODO check if policy allows requested to retrieve the block

	resp := &BlockResponse{
		RequestID: payload.RequestID,
		Block:     blockBytes,
	}

	signer := w.addressBook.GetLocalPeerKey()
	if err := w.Send(context.Background(), resp, payload.Signature.Key, blocks.SignWith(signer)); err != nil {
		w.logger.Warn("blx.handleRequestBlock could not send block", zap.Error(err))
		return err
	}

	return nil
}

func (w *exchange) Get(ctx context.Context, id string) (interface{}, error) {
	// Check local storage for block
	if blockBytes, err := w.store.Get(id); err == nil {
		return blocks.UnpackDecode(blockBytes)
	}

	req := &BlockRequest{
		RequestID: RandStringBytesMaskImprSrc(8),
		ID:        id,
		response:  make(chan interface{}),
	}

	defer func() {
		w.getRequests.Delete(req.RequestID)
		close(req.response)
	}()

	w.getRequests.Store(req.RequestID, req)
	signer := w.addressBook.GetLocalPeerKey()

	go func() {
		providers, err := w.discovery.GetProviders(ctx, id)
		if err != nil {
			// TODO log err
			return
		}

		for provider := range providers {
			if err := w.Send(ctx, req, provider, blocks.SignWith(signer)); err != nil {
				w.logger.Warn("blx.Get could not send req block", zap.Error(err))
			}
		}
	}()

	for {
		select {
		case payload := <-req.response:
			return payload, nil

		case <-ctx.Done():
			return nil, storage.ErrNotFound
		}
	}
}

func (w *exchange) Send(ctx context.Context, o blocks.Typed, recipient *crypto.Key, opts ...blocks.PackOption) error {
	// b := blocks.Encode(o)
	// if err := blocks.Sign(b, w.addressBook.GetLocalPeerInfo().Key); err != nil {
	// 	return err
	// }

	bytes, err := blocks.PackEncode(o, opts...)
	if err != nil {
		return err
	}

	if os.Getenv("DEBUG_BLOCKS") != "" {
		fmt.Print("> ---------- out block / start ")
		b, _ := json.MarshalIndent(o, "> ", "  ")
		fmt.Print(string(b))
		fmt.Println(" ---------- out block / end")
	}

	SendBlockEvent(
		"outgoing",
		o.GetType(),
		len(bytes),
	)

	if blocks.ShouldPersist(blocks.GetFromType(reflect.TypeOf(o))) {
		blockID, _ := blocks.SumSha3(bytes)
		w.store.Store(blockID, bytes)
	}

	// TODO right now there is no way to error on this, do we have to?
	w.outgoingPayloads <- outBlock{
		recipient: recipient,
		bytes:     bytes,
	}

	return nil
}

func (w *exchange) writeBlock(ctx context.Context, bytes []byte, rw io.ReadWriter) error {
	if _, err := rw.Write(bytes); err != nil {
		return err
	}

	w.logger.Debug("writing block", zap.String("bytes", base58.Encode(bytes)))
	return nil
}

func (w *exchange) getOrDialRelay(ctx context.Context, peerID string) (net.Conn, error) {
	peer, err := w.addressBook.GetPeerInfo(peerID)
	if err != nil {
		return nil, err
	}

	for _, address := range peer.Addresses {
		// TODO better check
		if strings.HasPrefix(address, "relay:") {
			relayPeerID := strings.Replace(address, "relay:", "", 1)
			conn, err := w.GetOrDial(ctx, relayPeerID)
			if err != nil {
				continue
			}
			return conn, nil
		}
	}

	return nil, ErrAllAddressesFailed
}

func (w *exchange) GetOrDial(ctx context.Context, peerID string) (net.Conn, error) {
	w.logger.Debug("looking for existing connection", zap.String("peer_id", peerID))
	if peerID == "" {
		return nil, errors.New("missing peer id")
	}

	existingConn, ok := w.streams.Load(peerID)
	if ok {
		return existingConn.(net.Conn), nil
	}

	conn, err := w.network.Dial(ctx, peerID)
	if err != nil {
		w.Close(peerID, conn)
		return nil, err
	}

	// TODO move after handshake
	// handle outgoing connections
	w.outgoing <- conn

	// store conn for reuse
	w.streams.Store(peerID, conn)

	w.logger.Debug("writing handshake")

	// handshake so the other side knows who we are
	// TODO can't ths use Send()?
	handshake := &HandshakePayload{
		PeerInfo: w.addressBook.GetLocalPeerInfo(),
	}

	signer := w.addressBook.GetLocalPeerKey()
	handshakeBytes, err := blocks.PackEncode(handshake, blocks.SignWith(signer))
	if err != nil {
		panic(err)
		return nil, err
	}

	if err := w.writeBlock(ctx, handshakeBytes, conn); err != nil {
		w.Close(peerID, conn)
		panic(err)
		return nil, err
	}

	return conn, nil
}

// Listen on an address
// TODO do we need to return a listener?
func (w *exchange) Listen(ctx context.Context, addr string) (net.Listener, error) {
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
					if err := w.HandleConnection(conn); err != nil {
						w.logger.Warn("failed to handle block", zap.Error(err))
					}
				}()
			case conn := <-w.outgoing:
				go func() {
					if err := w.HandleConnection(conn); err != nil {
						w.logger.Warn("failed to handle block", zap.Error(err))
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

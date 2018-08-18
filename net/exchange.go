package net

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ugorji/go/codec"
	"go.uber.org/zap"

	"github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/log"
	"github.com/nimona/go-nimona/peers"
	"github.com/nimona/go-nimona/storage"
	"github.com/nimona/go-nimona/utils"
)

var (
	// ErrMissingPeer    = errors.New("Peer ID missing")
	// ErrInvalidBlock   = errors.New("Invalid Block")
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

// BlockHandler to handle incoming blocks
type BlockHandler func(event *blocks.Block) error

// Exchange interface for mocking exchange
type Exchange interface {
	Get(ctx context.Context, id string) (*blocks.Block, error)
	GetLocalBlocks() ([]string, error)
	Handle(contentType string, h BlockHandler) error
	Send(ctx context.Context, block *blocks.Block, recipients ...string) error
	Listen(ctx context.Context, addrress string) (net.Listener, error)
	RegisterDiscoverer(discovery Discoverer)
	Verify(block *blocks.Block) error
	Sign(block *blocks.Block, signerPeerInfo *peers.PrivatePeerInfo) error
}

type exchange struct {
	network     Networker
	addressBook *peers.AddressBook
	discovery   Discoverer

	outgoingBlocks chan outBlock
	incoming       chan net.Conn
	outgoing       chan net.Conn
	close          chan bool

	streams    sync.Map
	handlers   []handler
	logger     *zap.Logger
	streamLock utils.Kmutex

	store         storage.Storage
	getRequests   sync.Map
	subscriptions sync.Map
}

type outBlock struct {
	recipient string
	block     *blocks.Block
}

type incBlock struct {
	peerID string
	conn   net.Conn
	block  *blocks.Block
}

type handler struct {
	contentType string
	handler     BlockHandler
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

		outgoingBlocks: make(chan outBlock, 100),
		incoming:       make(chan net.Conn),
		outgoing:       make(chan net.Conn),
		close:          make(chan bool),

		handlers:   []handler{},
		logger:     log.Logger(ctx).Named("exchange"),
		streamLock: utils.NewKmutex(),

		store:       store,
		getRequests: sync.Map{},
	}

	signer := w.addressBook.GetLocalPeerInfo()

	go func() {
		for block := range w.outgoingBlocks {
			if signer.ID == block.recipient {
				w.logger.Info("cannot send block to self",
					zap.String("blockID", blocks.BestEffortID(block.block)))
				continue
			}

			if block.block.Metadata.Signer == "" || block.block.Metadata.Signer == signer.ID {
				if err := w.Sign(block.block, signer); err != nil {
					// TODO log eror
				}
			}

			// TODO log error and reconsider the async
			// TODO also maybe we need to verify it or something?

			blockID, err := block.block.ID()
			if err != nil {
				// TODO log error
				// return err
				continue
			}

			if !block.block.Metadata.Ephemeral {
				go w.store.Store(blockID, block.block)
			}

			logger := w.logger.With(zap.String("peerID", block.recipient))

			// try to send the block directly to the recipient
			w.logger.Debug("getting conn to write block", zap.String("recipient", block.recipient))
			conn, err := w.GetOrDial(ctx, block.recipient)
			if err != nil {
				logger.Debug("could not get conn to recipient", zap.String("recipient", block.recipient), zap.Error(err))
			} else {
				if err := w.writeBlock(ctx, block.block, conn); err != nil {
					// TODO better handling of connection errors
					w.Close(block.recipient, conn)
					logger.Debug("could not write to recipient", zap.Error(err), zap.String("recipient", block.recipient))
				} else {
					// update peer status
					w.addressBook.PutPeerStatus(block.recipient, peers.StatusConnected)
					continue
				}
			}

			// else try to send message via their relay addresses
			conn, err = w.getOrDialRelay(ctx, block.recipient)
			if err != nil {
				logger.Debug("could not get conn to recipient's relay", zap.String("recipient", block.recipient), zap.Error(err))
				continue
			}

			// create forwarded block
			fwBlock := blocks.NewEphemeralBlock(TypeForwarded, PayloadForwarded{
				RecipientID: block.recipient,
				Block:       block.block,
			})
			if err := w.Sign(fwBlock, signer); err != nil {
				// TODO log eror
				continue
			}

			// try to send the block directly to the recipient
			if err := w.writeBlock(ctx, fwBlock, conn); err != nil {
				// TODO better handling of connection errors
				// TODO this is a bad close, id is of recipient, conn is of relay
				w.Close(block.recipient, conn)
				logger.Debug("could not write to relay", zap.Error(err), zap.String("recipient", block.recipient))
				// update peer status
				w.addressBook.PutPeerStatus(block.recipient, peers.StatusError)
				continue
			}

			// update peer status
			w.addressBook.PutPeerStatus(block.recipient, peers.StatusCanConnect)
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

func (w *exchange) Handle(contentType string, h BlockHandler) error {
	w.handlers = append(w.handlers, handler{
		contentType: contentType,
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

	blockDecoder := codec.NewDecoder(conn, blocks.CborHandler())
	for {
		block := &blocks.Block{}
		if err := blockDecoder.Decode(block); err != nil {
			w.logger.Error("could not read block", zap.Error(err))
			w.Close("", conn)
			return err
		}

		if err := w.Process(block, conn); err != nil {
			w.Close("", conn)
			return err
		}
	}
}

// Process incoming block
func (w *exchange) Process(block *blocks.Block, conn net.Conn) error {
	if os.Getenv("DEBUG_BLOCKS") != "" {
		fmt.Println("< ---------- inc block / start")
		b, _ := json.MarshalIndent(block, "< ", "  ")
		fmt.Println(string(b))
		fmt.Println("< ---------- inc block / end")
	}

	if err := w.Verify(block); err != nil {
		w.logger.Warn("could not verify block", zap.Error(err))
		return err
	}

	eb, _ := blocks.Marshal(block)
	tb, _ := blocks.Marshal(block.Payload)
	SendBlockEvent(
		false,
		block.Type,
		0, // len(GetRecipientsFromBlockPolicies(block)),
		len(tb),
		len(eb),
	)

	blockID, err := block.ID()
	if err != nil {
		return err
	}

	if !block.Metadata.Ephemeral {
		if err := w.store.Store(blockID, block); err != nil {
			if err != storage.ErrExists {
				w.logger.Warn("could not write block", zap.Error(err))
			}
		}
	}

	contentType := block.Type

	if block.GetHeader("requestID") != "" {
		if err := w.handleTransferBlock(block); err != nil {
			w.logger.Warn("could not handle transfer block", zap.Error(err))
		}
	}

	// TODO convert these into proper handlers
	switch payload := block.Payload.(type) {
	case PayloadForwarded:
		w.logger.Info("got forwarded message", zap.String("recipient", payload.RecipientID))
		w.outgoingBlocks <- outBlock{
			recipient: payload.RecipientID,
			block:     payload.Block,
		}
		return nil

	case BlockRequest:
		if err := w.handleRequestBlock(block); err != nil {
			w.logger.Warn("could not handle request block", zap.Error(err))
		}

	case HandshakeBlock:
		if err := w.addressBook.PutPeerInfoFromBlock(payload.PeerInfo); err != nil {
			return err
		}

		w.streams.Store(block.Metadata.Signer, conn)
		return nil
	}

	for _, handler := range w.handlers {
		if strings.HasPrefix(contentType, handler.contentType) {
			if err := handler.handler(block); err != nil {
				w.logger.Info(
					"Could not handle event",
					zap.String("contentType", contentType),
					zap.Error(err),
				)
				return err
			}
		}
	}

	return nil
}

func (w *exchange) handleTransferBlock(block *blocks.Block) error {
	// Check if nonce exists in local addressBook
	value, ok := w.getRequests.Load(block.GetHeader("requestID"))
	if !ok {
		return nil
	}

	req, ok := value.(*BlockRequest)
	if !ok {
		return ErrInvalidRequest
	}

	req.response <- block

	return nil
}

func (w *exchange) handleRequestBlock(incBlock *blocks.Block) error {
	payload := incBlock.Payload.(BlockRequest)
	block, err := w.store.Get(payload.ID)
	if err != nil {
		return err
	}

	// TODO check if policy allows requested to retrieve the block
	block.SetHeader("requestID", payload.RequestID)
	if err := w.Send(context.Background(), block, incBlock.Metadata.Signer); err != nil {
		w.logger.Warn("blx.handleRequestBlock could not send block", zap.Error(err))
		return err
	}

	return nil
}

func (w *exchange) Get(ctx context.Context, id string) (*blocks.Block, error) {
	// Check local storage for block
	if block, err := w.store.Get(id); err == nil {
		return block, nil
	}

	req := &BlockRequest{
		RequestID: RandStringBytesMaskImprSrc(8),
		ID:        id,
		response:  make(chan *blocks.Block),
	}

	defer close(req.response)

	w.getRequests.Store(req.RequestID, req)

	go func() {
		providers, err := w.discovery.GetProviders(ctx, id)
		if err != nil {
			// TODO log err
			return
		}

		reqBlock := blocks.NewEphemeralBlock(BlockRequestType, req)
		for provider := range providers {
			if err := w.Send(ctx, reqBlock, provider); err != nil {
				w.logger.Warn("blx.Get could not send req block", zap.Error(err))
			}
		}
	}()

	for {
		select {
		case block := <-req.response:
			return block, nil

		case <-ctx.Done():
			return nil, storage.ErrNotFound
		}
	}
}

func (w *exchange) Send(ctx context.Context, block *blocks.Block, recipients ...string) error {
	// TODO do we need to send this to the policy recipients as well?
	// recipients = append(recipients, GetRecipientsFromBlockPolicies(block)...)
	for _, recipient := range recipients {
		// TODO right now there is no way to error on this, do we have to?
		w.outgoingBlocks <- outBlock{
			recipient: recipient,
			block:     block,
		}
	}

	return nil
}

func (w *exchange) GetLocalBlocks() ([]string, error) {
	return w.store.List()
}

func (w *exchange) writeBlock(ctx context.Context, block *blocks.Block, rw io.ReadWriter) error {
	if os.Getenv("DEBUG_BLOCKS") != "" {
		fmt.Println("> ---------- out block / start")
		b, _ := json.MarshalIndent(block, "> ", "  ")
		fmt.Println(string(b))
		fmt.Println("> ---------- out block / end")
	}

	blockBytes, err := blocks.Marshal(block)
	if err != nil {
		return err
	}

	if _, err := rw.Write(blockBytes); err != nil {
		return err
	}

	tb, _ := blocks.Marshal(block.Payload)
	SendBlockEvent(
		true,
		block.Type,
		0, // len(GetRecipientsFromBlockPolicies(block)),
		len(tb),
		len(blockBytes),
	)

	w.logger.Debug("writing block", zap.Any("block", block))

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
	handshakeBlock := blocks.NewEphemeralBlock(
		"handshake",
		HandshakeBlock{
			PeerInfo: w.addressBook.GetLocalPeerInfo().Block(),
		},
		peerID,
	)
	signer := w.addressBook.GetLocalPeerInfo()
	if err := w.Sign(handshakeBlock, signer); err != nil {
		return nil, err
	}

	if err := w.writeBlock(ctx, handshakeBlock, conn); err != nil {
		w.Close(peerID, conn)
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

// Sign block given a private peer info
// TODO signer should already be set in the block, so maybe we can get
// the keys from the address book?
func (w *exchange) Sign(block *blocks.Block, signerPeerInfo *peers.PrivatePeerInfo) error {
	block.Metadata.Signer = signerPeerInfo.ID
	digest, err := blocks.GetSignatureDigest(block)
	if err != nil {
		return err
	}

	signature, err := peers.SignData(digest, signerPeerInfo.PrivateKey)
	if err != nil {
		return err
	}

	block.Signature = signature
	return nil
}

// Verify block's signature
func (w *exchange) Verify(block *blocks.Block) error {
	if len(block.Signature) == 0 {
		return nil
	}

	digest, err := blocks.GetSignatureDigest(block)
	if err != nil {
		return err
	}

	return peers.Verify(block.Metadata.Signer, digest, block.Signature)
}

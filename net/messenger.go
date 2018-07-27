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

	"github.com/nimona/go-nimona/log"
	"github.com/nimona/go-nimona/utils"
)

var (
	// ErrMissingPeer    = errors.New("Peer ID missing")
	// ErrInvalidBlock   = errors.New("Invalid Block")
	ErrInvalidRequest = errors.New("Invalid request")
)

const (
	TypeHandshake = "handshake"
	// PayloadTypeTransferBlock type for PayloadTransferBlock
	// PayloadTypeTransferBlock = "blx.transfer-block"
	// PayloadTypeRequestBlock type for PayloadRequestBlock
	PayloadTypeRequestBlock = "blx.request-block"
)

func init() {
	RegisterContentType(TypeHandshake, HandshakeBlock{})
	// RegisterContentType(PayloadTypeTransferBlock, PayloadTransferBlock{})
	RegisterContentType(PayloadTypeRequestBlock, PayloadRequestBlock{})
}

var (
	// ErrAllAddressesFailed for when a peer cannot be dialed
	ErrAllAddressesFailed = errors.New("all addresses failed to dial")
	// ErrNotForUs block is not meant for us
	ErrNotForUs = errors.New("block not for us")
)

// BlockHandler to handle incoming blocks
type BlockHandler func(event *Block) error

// Messenger interface for mocking messenger
type Messenger interface {
	Get(ctx context.Context, id string) (*Block, error)
	GetLocalBlocks() ([]string, error)
	Handle(contentType string, h BlockHandler) error
	Send(ctx context.Context, block *Block, recipients ...string) error
	Listen(ctx context.Context, addrress string) (net.Listener, error)
	RegisterDiscoverer(discovery Discoverer)
}

type messenger struct {
	network     Networker
	addressBook PeerManager
	discovery   Discoverer

	// incomingBlocks chan *Block
	outgoingBlocks chan outBlock
	incoming       chan net.Conn
	outgoing       chan net.Conn
	close          chan bool

	streams    sync.Map
	handlers   []handler
	logger     *zap.Logger
	streamLock utils.Kmutex

	storage       Storage
	getRequests   sync.Map
	subscriptions sync.Map
}

type outBlock struct {
	recipient string
	block     *Block
}

type incBlock struct {
	peerID string
	conn   net.Conn
	block  *Block
}

type handler struct {
	contentType string
	handler     BlockHandler
}

// NewMessenger creates a messenger on a given network
func NewMessenger(addressBook *AddressBook, storage Storage) (Messenger, error) {
	ctx := context.Background()

	network, err := NewNetwork(addressBook)
	if err != nil {
		return nil, err
	}

	w := &messenger{
		network:     network,
		addressBook: addressBook,

		// incomingBlocks: make(chan *Block, 100),
		outgoingBlocks: make(chan outBlock, 100),
		incoming:       make(chan net.Conn),
		outgoing:       make(chan net.Conn),
		close:          make(chan bool),

		handlers:   []handler{},
		logger:     log.Logger(ctx).Named("messenger"),
		streamLock: utils.NewKmutex(),

		storage:     storage,
		getRequests: sync.Map{},
	}

	signer := w.addressBook.GetLocalPeerInfo()

	go func() {
		for block := range w.outgoingBlocks {
			if signer.ID == block.recipient {
				continue
			}

			if block.block.Metadata.Signer == "" || block.block.Metadata.Signer == signer.ID {
				SetSigner(block.block, signer)
				if err := Sign(block.block, signer); err != nil {
					// TODO log eror
				}
				if err := SetID(block.block); err != nil {
					// TODO log error
				}
			}

			// TODO log error and reconsider the async
			// TODO also maybe we need to verify it or something?
			if block.block.Metadata.ID != "" && !block.block.Metadata.Ephemeral {
				go w.storage.Store(block.block.Metadata.ID, block.block)
			}

			logger := w.logger.With(zap.String("peerID", block.recipient))

			// w.streamLock.Lock(recipient)
			// defer w.streamLock.Unlock(recipient)

			w.logger.Debug("getting conn to write block", zap.String("recipient", block.recipient))
			conn, err := w.GetOrDial(ctx, block.recipient)
			if err != nil {
				logger.Debug("could not get conn to recipient", zap.String("recipient", block.recipient), zap.Error(err))
				continue
			}

			// TODO this seems messy
			// try to send the block directly to the recipient
			if err := w.writeBlock(ctx, block.block, conn); err != nil {
				w.Close(block.recipient, conn)
				logger.Debug("could not send directly to recipient", zap.Error(err), zap.String("recipient", block.recipient))
				continue
			}
		}
	}()

	return w, nil
}

func (w *messenger) RegisterDiscoverer(discovery Discoverer) {
	w.discovery = discovery

	ctx := context.Background()
	go func() {
		for {
			blocks, err := w.storage.List()
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

func (w *messenger) Handle(contentType string, h BlockHandler) error {
	w.handlers = append(w.handlers, handler{
		contentType: contentType,
		handler:     h,
	})
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

	blockDecoder := codec.NewDecoder(conn, getCborHandler())
	for {
		block := &Block{}
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
func (w *messenger) Process(block *Block, conn net.Conn) error {
	if err := block.Verify(); err != nil {
		w.logger.Warn("could not verify block", zap.Error(err))
		return err
	}

	eb, _ := Marshal(block)
	tb, _ := Marshal(block.Payload)
	SendBlockEvent(
		false,
		block.Metadata.Type,
		len(GetRecipientsFromBlockPolicies(block)),
		len(tb),
		len(eb),
	)

	if os.Getenv("DEBUG_BLOCKS") != "" {
		fmt.Println("< ---------- inc block / start")
		b, _ := json.MarshalIndent(block, "< ", "  ")
		fmt.Println(string(b))
		fmt.Println("< ---------- inc block / end")
	}

	if !block.Metadata.Ephemeral && block.Metadata.ID != "" {
		if err := w.storage.Store(block.Metadata.ID, block); err != nil {
			if err != ErrExists {
				w.logger.Warn("could not write block", zap.Error(err))
			}
		}
	}

	contentType := block.Metadata.Type

	if block.GetHeader("requestID") != "" {
		if err := w.handleTransferBlock(block); err != nil {
			w.logger.Warn("could not handle transfer block", zap.Error(err))
		}
	}

	switch contentType {
	case PayloadTypeRequestBlock:
		if err := w.handleRequestBlock(block); err != nil {
			w.logger.Warn("could not handle request block", zap.Error(err))
		}

	case TypeHandshake:
		// TODO move to handleHandshake
		payload, ok := block.Payload.(HandshakeBlock)
		if !ok {
			return errors.New("payload does not match expected handshake")
		}

		if err := w.addressBook.PutPeerInfoFromBlock(payload.PeerInfo); err != nil {
			return err
		}

		w.streams.Store(block.Metadata.Signer, conn)
		return nil
	}

	var blockHandler BlockHandler
	ok := false
	for _, handler := range w.handlers {
		if !strings.HasPrefix(contentType, handler.contentType) {
			continue
		}
		ok = true
		blockHandler = handler.handler
		break
	}

	if !ok {
		w.logger.Info(
			"No handler registered for contentType",
			zap.String("contentType", contentType),
		)
		return nil
	}

	if err := blockHandler(block); err != nil {
		w.logger.Info(
			"Could not handle event",
			zap.String("contentType", contentType),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (w *messenger) handleTransferBlock(block *Block) error {
	// Check if nonce exists in local addressBook
	value, ok := w.getRequests.Load(block.GetHeader("requestID"))
	if !ok {
		return nil
	}

	req, ok := value.(*PayloadRequestBlock)
	if !ok {
		return ErrInvalidRequest
	}

	req.response <- block

	return nil
}

func (w *messenger) handleRequestBlock(incBlock *Block) error {
	payload := incBlock.Payload.(PayloadRequestBlock)
	block, err := w.storage.Get(payload.ID)
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

func (w *messenger) Get(ctx context.Context, id string) (*Block, error) {
	// Check local storage for block
	if block, err := w.storage.Get(id); err == nil {
		return block, nil
	}

	req := &PayloadRequestBlock{
		RequestID: RandStringBytesMaskImprSrc(8),
		ID:        id,
		response:  make(chan *Block),
	}

	defer close(req.response)

	w.getRequests.Store(req.RequestID, req)

	providers, err := w.discovery.GetProviders(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	// Request block
	wrapper := NewEphemeralBlock(PayloadTypeRequestBlock, req)
	if err := w.Send(ctx, wrapper, providers...); err != nil {
		w.logger.Warn("blx.Get could not send block", zap.Error(err))
		return nil, err
	}

	for {
		select {
		case block := <-req.response:
			return block, nil

		case <-ctx.Done():
			return nil, ErrNotFound
		}
	}
}

func (w *messenger) Send(ctx context.Context, block *Block, recipients ...string) error {
	// TODO do this async or after sending?
	if !block.Metadata.Ephemeral && block.Metadata.ID != "" {
		if err := w.storage.Store(block.Metadata.ID, block); err != nil {
			if err != ErrExists {
				w.logger.Warn("could not write block", zap.Error(err))
			}
		}
	}
	// TODO do we need to send this to the policy recipients as well?
	recipients = append(recipients, GetRecipientsFromBlockPolicies(block)...)
	for _, recipient := range recipients {
		// TODO deal with error
		// if err := w.sendOne(ctx, block, recipient); err != nil {
		w.outgoingBlocks <- outBlock{
			recipient: recipient,
			block:     block,
		}
		// w.logger.Warn("could not send block", zap.Error(err), zap.String("recipient", recipient))
		// }
	}

	return nil
}

func (w *messenger) GetLocalBlocks() ([]string, error) {
	return w.storage.List()
}

// func (w *messenger) sendOne(ctx context.Context, block *Block, recipient string) error {
// 	logger := w.logger.With(zap.String("peerID", recipient))

// 	w.streamLock.Lock(recipient)
// 	defer w.streamLock.Unlock(recipient)

// 	w.logger.Debug("getting conn to write block", zap.String("recipient", recipient))
// 	conn, err := w.GetOrDial(ctx, recipient)
// 	if err != nil {
// 		return err
// 	}

// 	// TODO this seems messy
// 	// try to send the block directly to the recipient
// 	if err := w.writeBlock(ctx, block, conn); err != nil {
// 		w.Close(recipient, conn)
// 		logger.Debug("could not send directly to recipient", zap.Error(err))
// 	} else {
// 		return nil
// 	}

// 	return ErrAllAddressesFailed
// }

func (w *messenger) writeBlock(ctx context.Context, block *Block, rw io.ReadWriter) error {
	if os.Getenv("DEBUG_BLOCKS") != "" {
		fmt.Println("> ---------- out block / start")
		b, _ := json.MarshalIndent(block, "> ", "  ")
		fmt.Println(string(b))
		fmt.Println("> ---------- out block / end")
	}

	blockBytes, err := Marshal(block)
	if err != nil {
		return err
	}

	if _, err := rw.Write(blockBytes); err != nil {
		return err
	}

	tb, _ := Marshal(block.Payload)
	SendBlockEvent(
		true,
		block.Metadata.Type,
		len(GetRecipientsFromBlockPolicies(block)),
		len(tb),
		len(blockBytes),
	)

	w.logger.Debug("writing block", zap.Any("block", block))

	return nil
}

func (w *messenger) GetOrDial(ctx context.Context, peerID string) (net.Conn, error) {
	w.logger.Debug("getting conn", zap.String("peer_id", peerID))
	if peerID == "" {
		return nil, errors.New("missing peer id")
	}

	existingConn, ok := w.streams.Load(peerID)
	if ok {
		w.addressBook.PutPeerStatus(peerID, StatusConnected)
		return existingConn.(net.Conn), nil
	}

	w.logger.Debug("dialing peer", zap.String("peer_id", peerID))
	w.addressBook.PutPeerStatus(peerID, StatusConnecting)
	conn, err := w.network.Dial(ctx, peerID)
	if err != nil {
		w.addressBook.PutPeerStatus(peerID, StatusError)
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
	handshakeBlock := NewEphemeralBlock(
		"handshake",
		HandshakeBlock{
			PeerInfo: w.addressBook.GetLocalPeerInfo().Block(),
		},
		peerID,
	)
	signer := w.addressBook.GetLocalPeerInfo()
	SetSigner(handshakeBlock, signer)
	if err := Sign(handshakeBlock, signer); err != nil {
		return nil, err
	}

	if err := SetID(handshakeBlock); err != nil {
		return nil, err
	}

	if err := w.writeBlock(ctx, handshakeBlock, conn); err != nil {
		w.addressBook.PutPeerStatus(peerID, StatusError)
		w.Close(peerID, conn)
		return nil, err
	}

	w.addressBook.PutPeerStatus(peerID, StatusConnected)
	return conn, nil
}

// Listen on an address
// TODO do we need to return a listener?
func (w *messenger) Listen(ctx context.Context, addr string) (net.Listener, error) {
	if w.discovery == nil {
		return nil, errors.New("missing discoverer")
	}

	listener, err := w.network.Listen(ctx, addr)
	if err != nil {
		return nil, err
	}

	closed := false

	go func() {
		for {
			select {
			// case block := <-w.incomingBlocks:
			// 	if err := w.Process(block, nil); err != nil {
			// 		w.logger.Warn("failed to process block", zap.Error(err))
			// 	}
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

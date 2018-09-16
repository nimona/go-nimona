package net

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"go.uber.org/zap"

	"nimona.io/go/blocks"
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
	RegisterDiscoverer(discovery Discoverer)
}

type exchange struct {
	network     Networker
	addressBook *peers.AddressBook
	manager     *ConnectionManager
	discovery   Discoverer

	outgoingPayloads chan *outBlock
	incomingPayloads chan *incBlock

	handlers   []handler
	logger     *zap.Logger
	streamLock utils.Kmutex

	store         storage.Storage
	getRequests   sync.Map
	subscriptions sync.Map
}

type outBlock struct {
	context   context.Context
	recipient *crypto.Key
	opts      []blocks.PackOption
	typed     blocks.Typed
	err       chan error
}

type incBlock struct {
	conn  *Connection
	typed blocks.Typed
}

type handler struct {
	contentType glob.Glob
	handler     func(o blocks.Typed) error
}

// NewExchange creates a exchange on a given network
func NewExchange(addressBook *peers.AddressBook, store storage.Storage, address string) (Exchange, error) {
	ctx := context.Background()

	network, err := NewNetwork(addressBook)
	if err != nil {
		return nil, err
	}

	w := &exchange{
		network:     network,
		addressBook: addressBook,
		manager:     &ConnectionManager{},

		outgoingPayloads: make(chan *outBlock, 10),
		incomingPayloads: make(chan *incBlock, 10),

		handlers:   []handler{},
		logger:     log.Logger(ctx).Named("exchange"),
		streamLock: utils.NewKmutex(),

		store:       store,
		getRequests: sync.Map{},
	}

	self := w.addressBook.GetLocalPeerInfo()
	incomingConnections, err := w.network.Listen(ctx, address)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn := <-incomingConnections
			go func(conn *Connection) {
				w.manager.Add(conn)
				if err := w.HandleConnection(conn); err != nil {
					w.logger.Warn("failed to handle block", zap.Error(err))
				}
			}(conn)
		}
	}()

	go func() {
		for {
			block := <-w.incomingPayloads
			go func(block *incBlock) {
				if err := w.process(block.typed, block.conn); err != nil {
					w.logger.Error("getting conn to write block", zap.Error(err))
				}
			}(block)
		}
	}()

	go func() {
		for {
			block := <-w.outgoingPayloads
			recipientThumbPrint := block.recipient.Thumbprint()
			if self.Thumbprint() == recipientThumbPrint {
				w.logger.Info("cannot send block to self")
				block.err <- errors.New("cannot send block to self")
				continue
			}

			logger := w.logger.With(zap.String("peerID", recipientThumbPrint))

			// try to send the block directly to the recipient
			logger.Debug("getting conn to write block")
			conn, err := w.GetOrDial(ctx, recipientThumbPrint)
			if err != nil {
				logger.Debug("could not get conn to recipient", zap.Error(err))
				block.err <- err
				continue
			}

			if err := Write(block.typed, conn, block.opts...); err != nil {
				w.manager.Close(recipientThumbPrint)
				logger.Debug("could not write to recipient", zap.Error(err))
				block.err <- err
				continue
			}

			// update peer status
			w.addressBook.PutPeerStatus(recipientThumbPrint, peers.StatusConnected)
			block.err <- nil
			continue
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

func (w *exchange) HandleConnection(conn *Connection) error {
	w.logger.Debug("handling new connection", zap.String("remote", conn.RemoteID))
	for {
		// TODO use decoder
		typed, err := Read(conn)
		if err != nil {
			return err
		}

		w.incomingPayloads <- &incBlock{
			conn:  conn,
			typed: typed,
		}
	}
}

// Process incoming block
func (w *exchange) process(typed blocks.Typed, conn *Connection) error {
	// TODO verify signature

	// TODO convert these into proper handlers
	switch payload := typed.(type) {
	case *ForwardRequest:
		w.logger.Info("got forwarded message", zap.String("recipient", payload.Recipient.Thumbprint()))
		w.outgoingPayloads <- &outBlock{
			recipient: payload.Recipient,
			typed:     payload.Typed,
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
	}

	contentType := typed.GetType()
	for _, handler := range w.handlers {
		if handler.contentType.Match(contentType) {
			if err := handler.handler(typed); err != nil {
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

func (w *exchange) Send(ctx context.Context, typed blocks.Typed, recipient *crypto.Key, opts ...blocks.PackOption) error {
	bytes, err := blocks.PackEncode(typed, opts...)
	if err != nil {
		return err
	}

	SendBlockEvent(
		"outgoing",
		typed.GetType(),
		len(bytes),
	)

	if blocks.ShouldPersist(typed.GetType()) {
		blockID, _ := blocks.SumSha3(bytes)
		w.store.Store(blockID, bytes)
	}

	cerr := make(chan error)

	w.outgoingPayloads <- &outBlock{
		context:   ctx,
		recipient: recipient,
		typed:     typed,
		err:       cerr,
		opts:      opts,
	}

	var sErr error
	select {
	case err := <-cerr:
		sErr = err
	case <-time.After(time.Second * 5):
		sErr = errors.New("giving up, timing out")
	}

	return sErr

	// TODO fix forwarding

	// if sErr == nil {
	// 	fmt.Println("---------------- sErr", sErr)
	// 	return nil
	// }

	// // try to send message via their relay addresses
	// peerID := recipient.Thumbprint()
	// peer, err := w.addressBook.GetPeerInfo(peerID)
	// if err != nil {
	// 	return ErrAllAddressesFailed
	// }

	// // we unpack this because of the Send opts that might have signed this
	// fwTyped, err := blocks.UnpackDecode(bytes)
	// if err != nil {
	// 	return err
	// }

	// fw := &ForwardRequest{
	// 	Recipient: recipient,
	// 	Typed:     fwTyped,
	// }

	// for _, address := range peer.Addresses {
	// 	if strings.HasPrefix(address, "relay:") {
	// 		relayPeerID := strings.Replace(address, "relay:", "", 1)
	// 		relayPeer, err := w.addressBook.GetPeerInfo(relayPeerID)
	// 		if err != nil {
	// 			continue
	// 		}
	// 		if err := w.Send(ctx, fw, relayPeer.Signature.Key); err == nil {
	// 			return nil
	// 		}
	// 	}
	// }

	// return ErrAllAddressesFailed
}

func (w *exchange) GetOrDial(ctx context.Context, peerID string) (*Connection, error) {
	w.logger.Debug("looking for existing connection", zap.String("peer_id", peerID))
	if peerID == "" {
		return nil, errors.New("missing peer id")
	}

	existingConn, err := w.manager.Get(peerID)
	if err == nil {
		w.logger.Debug("found existing connection", zap.String("peerID", peerID))
		return existingConn, nil
	}

	conn, err := w.network.Dial(ctx, peerID)
	if err != nil {
		// w.manager.Close(peerID)
		return nil, err
	}

	go func() {
		if err := w.HandleConnection(conn); err != nil {
			w.logger.Warn("failed to handle block", zap.Error(err))
		}
	}()

	w.manager.Add(conn)

	return conn, nil
}

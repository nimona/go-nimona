package net

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"go.uber.org/zap"

	"nimona.io/go/log"
	"nimona.io/go/peers"
	"nimona.io/go/primitives"
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
	Handle(contentType string, h func(o *primitives.Block) error) (func(), error)
	Send(ctx context.Context, o *primitives.Block, recipient *primitives.Key, opts ...primitives.SendOption) error
	RegisterDiscoverer(discovery Discoverer)
}

type exchange struct {
	network     Networker
	addressBook *peers.AddressBook
	manager     *ConnectionManager
	discovery   Discoverer

	outgoingPayloads chan *outBlock
	incomingPayloads chan *incBlock

	handlers   sync.Map
	logger     *zap.Logger
	streamLock utils.Kmutex

	store         storage.Storage
	getRequests   sync.Map
	subscriptions sync.Map
}

type outBlock struct {
	context   context.Context
	recipient *primitives.Key
	opts      []primitives.SendOption
	block     *primitives.Block
	err       chan error
}

type incBlock struct {
	conn  *Connection
	typed *primitives.Block
}

type handler struct {
	contentType glob.Glob
	handler     func(o *primitives.Block) error
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

		handlers:   sync.Map{},
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
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("Recovered while processing", zap.Any("r", r))
					}
				}()
				if err := w.process(block.typed, block.conn); err != nil {
					w.logger.Error("processing block", zap.Error(err))
				}
			}(block)
		}
	}()

	go func() {
		for {
			block := <-w.outgoingPayloads
			if block.recipient == nil {
				w.logger.Info("missing recipient")
				block.err <- errors.New("missing recipient")
				continue
			}

			recipientThumbPrint := block.recipient.Thumbprint()
			if self.Thumbprint() == recipientThumbPrint {
				w.logger.Info("cannot send block to self")
				block.err <- errors.New("cannot send block to self")
				continue
			}

			logger := w.logger.With(zap.String("peerID", recipientThumbPrint))

			// try to send the block directly to the recipient
			logger.Debug("getting conn to write block")
			conn, err := w.GetOrDial(block.context, recipientThumbPrint)
			if err != nil {
				logger.Debug("could not get conn to recipient", zap.Error(err))
				block.err <- err
				continue
			}

			if err := Write(block.block, conn, block.opts...); err != nil {
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

func (w *exchange) Handle(typePatern string, h func(o *primitives.Block) error) (func(), error) {
	g, err := glob.Compile(typePatern, '.', '/', '#')
	if err != nil {
		return nil, err
	}
	hID := RandStringBytesMaskImprSrc(8)
	w.handlers.Store(hID, &handler{
		contentType: g,
		handler:     h,
	})
	r := func() {
		w.handlers.Delete(hID)
	}
	return r, nil
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
func (w *exchange) process(block *primitives.Block, conn *Connection) error {
	// TODO verify signature

	// TODO convert these into proper handlers
	switch block.Type {
	case "nimona.io/block.forward.request":
		req := &ForwardRequest{}
		req.FromBlock(block)
		w.logger.Info("got forwarded message", zap.String("recipient", req.Recipient.Thumbprint()))
		cerr := make(chan error, 1)
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		defer cf()
		w.outgoingPayloads <- &outBlock{
			context:   ctx,
			recipient: req.Recipient,
			block:     req.FwBlock,
			err:       cerr,
		}
		return <-cerr

	case "nimona.io/block.request":
		req := &BlockRequest{}
		req.FromBlock(block)
		if err := w.handleBlockRequest(req); err != nil {
			w.logger.Warn("could not handle request block", zap.Error(err))
		}

	case "nimona.io/block.response":
		res := &BlockResponse{}
		res.FromBlock(block)
		if err := w.handleBlockResponse(res); err != nil {
			w.logger.Warn("could not handle request block", zap.Error(err))
		}
	}

	hp := 0
	w.handlers.Range(func(_, v interface{}) bool {
		hp++
		h := v.(*handler)
		if h.contentType.Match(block.Type) {
			go func(h *handler, block *primitives.Block) {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("Recovered while handling", zap.Any("r", r))
					}
				}()
				if err := h.handler(block); err != nil {
					w.logger.Info(
						"Could not handle event",
						zap.String("contentType", block.Type),
						zap.Error(err),
					)
				}
			}(h, block)
		}
		return true
	})

	if hp == 0 {
		fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		fmt.Println("+++++ NO HANDLERS ++++++++++++++++++++++++++++++++++++++")
		fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		fmt.Println("+++++", block.Type)
		fmt.Println("+++++", block.Payload)
		fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	} else {
		if shouldPersist(block.Type) {
			bytes, _ := primitives.Marshal(block)
			w.store.Store(block.ID(), bytes)
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

	req.response <- payload.Block
	return nil
}

func (w *exchange) handleBlockRequest(payload *BlockRequest) error {
	blockBytes, err := w.store.Get(payload.ID)
	if err != nil {
		return err
	}

	// TODO check if policy allows requested to retrieve the block

	block, err := primitives.Unmarshal(blockBytes)
	if err != nil {
		return err
	}

	resp := &BlockResponse{
		RequestID:      payload.RequestID,
		RequestedBlock: block,
	}

	signer := w.addressBook.GetLocalPeerKey()
	if err := w.Send(context.Background(), resp.Block(), payload.Sender, primitives.SignWith(signer)); err != nil {
		w.logger.Warn("blx.handleBlockRequest could not send block", zap.Error(err))
		return err
	}

	return nil
}

func (w *exchange) Get(ctx context.Context, id string) (interface{}, error) {
	// Check local storage for block
	if blockBytes, err := w.store.Get(id); err == nil {
		block, err := primitives.Unmarshal(blockBytes)
		if err != nil {
			return nil, err
		}
		return block, nil
	}

	req := &BlockRequest{
		RequestID: RandStringBytesMaskImprSrc(8),
		ID:        id,
		response:  make(chan interface{}, 10),
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
			if err := w.Send(ctx, req.Block(), provider, primitives.SignWith(signer)); err != nil {
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

func (w *exchange) Send(ctx context.Context, block *primitives.Block, recipient *primitives.Key, opts ...primitives.SendOption) error {
	cfg := primitives.ParseSendOptions(opts...)

	if cfg.Sign && cfg.Key != nil && block.Signature == nil {
		if err := primitives.Sign(block, cfg.Key); err != nil {
			return err
		}
	}

	if shouldPersist(block.Type) {
		bytes, _ := primitives.Marshal(block)
		w.store.Store(block.ID(), bytes)
	}

	cerr := make(chan error, 1)
	w.outgoingPayloads <- &outBlock{
		context:   ctx,
		recipient: recipient,
		block:     block,
		err:       cerr,
		opts:      opts,
	}
	if err := <-cerr; err == nil {
		return nil
	}

	// try to send message via their relay addresses
	peerID := recipient.Thumbprint()
	peer, err := w.addressBook.GetPeerInfo(peerID)
	if err != nil {
		return err
	}

	// we unpack this because of the Send opts that might have signed this
	// fwTyped, err := primitives.UnpackDecode(bytes)
	// if err != nil {
	// 	return err
	// }

	// TODO(geoah) make sure fw block is signed
	fw := &ForwardRequest{
		Recipient: recipient,
		FwBlock:   block,
	}

	for _, address := range peer.Addresses {
		if strings.HasPrefix(address, "relay:") {
			relayPeerID := strings.Replace(address, "relay:", "", 1)
			if w.addressBook.GetLocalPeerInfo().Thumbprint() == relayPeerID {
				continue
			}
			relayPeer, err := w.addressBook.GetPeerInfo(relayPeerID)
			if err != nil {
				continue
			}
			cerr := make(chan error, 1)
			w.outgoingPayloads <- &outBlock{
				context:   ctx,
				recipient: relayPeer.Signature.Key,
				block:     fw.Block(),
				err:       cerr,
				opts:      opts,
			}
			if err := <-cerr; err == nil {
				return nil
			}
		}
	}

	return ErrAllAddressesFailed
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

func shouldPersist(t string) bool {
	if strings.HasPrefix(t, "nimona.io/dht") ||
		strings.HasPrefix(t, "nimona.io/telemetry") ||
		strings.HasPrefix(t, "nimona.io/handshake") {
		return false
	}
	return true
}

package net

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	"nimona.io/go/log"
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
	Handle(contentType string, h func(o *encoding.Object) error) (func(), error)
	Send(ctx context.Context, o *encoding.Object, address string) error
}

type exchange struct {
	key     *crypto.Key
	net     Network
	manager *ConnectionManager

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
	recipient string
	block     *encoding.Object
	err       chan error
}

type incBlock struct {
	conn  *Connection
	typed *encoding.Object
}

type handler struct {
	contentType glob.Glob
	handler     func(o *encoding.Object) error
}

// NewExchange creates a exchange on a given network
func NewExchange(key *crypto.Key, net Network, store storage.Storage, address string) (Exchange, error) {
	ctx := context.Background()

	w := &exchange{
		key:     key,
		net:     net,
		manager: &ConnectionManager{},

		outgoingPayloads: make(chan *outBlock, 10),
		incomingPayloads: make(chan *incBlock, 10),

		handlers:   sync.Map{},
		logger:     log.Logger(ctx).Named("exchange"),
		streamLock: utils.NewKmutex(),

		store:       store,
		getRequests: sync.Map{},
	}

	incomingConnections, err := w.net.Listen(ctx, address)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn := <-incomingConnections
			go func(conn *Connection) {
				w.manager.Add("peer:"+conn.RemoteID, conn)
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
			if block.recipient == "" {
				w.logger.Info("missing recipient")
				block.err <- errors.New("missing recipient")
				continue
			}

			logger := w.logger.With(zap.String("recipient", block.recipient))

			// try to send the block directly to the recipient
			logger.Debug("getting conn to write block")
			conn, err := w.GetOrDial(block.context, block.recipient)
			if err != nil {
				logger.Debug("could not get conn to recipient", zap.Error(err))
				block.err <- err
				continue
			}

			if err := Write(block.block, conn); err != nil {
				w.manager.Close(block.recipient)
				logger.Debug("could not write to recipient", zap.Error(err))
				block.err <- err
				continue
			}

			// update peer status
			// TODO(geoah) fix status -- not that we are using this
			// w.addressBook.PutPeerStatus(recipientThumbPrint, peers.StatusConnected)
			block.err <- nil
			continue
		}
	}()

	return w, nil
}

func (w *exchange) Handle(typePatern string, h func(o *encoding.Object) error) (func(), error) {
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

var (
	typeBlockForwardRequest = BlockForwardRequest{}.GetType()
	typeBlockRequest        = BlockRequest{}.GetType()
	typeBlockResponse       = BlockResponse{}.GetType()
)

// Process incoming block
func (w *exchange) process(o *encoding.Object, conn *Connection) error {
	// TODO verify signature
	// TODO convert these into proper handlers
	switch o.GetType() {
	case typeBlockForwardRequest:
		v := &BlockForwardRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		w.logger.Info("got forwarded message", zap.String("recipient", v.Recipient))
		cerr := make(chan error, 1)
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		defer cf()
		w.outgoingPayloads <- &outBlock{
			context:   ctx,
			recipient: v.Recipient,
			block:     v.FwBlock,
			err:       cerr,
		}
		return <-cerr

	case typeBlockRequest:
		v := &BlockRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		if err := w.handleBlockRequest(v); err != nil {
			w.logger.Warn("could not handle request block", zap.Error(err))
		}

	case typeBlockResponse:
		v := &BlockResponse{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		if err := w.handleBlockResponse(v); err != nil {
			w.logger.Warn("could not handle request block", zap.Error(err))
		}
	}

	ct := o.GetType()

	hp := 0
	w.handlers.Range(func(_, v interface{}) bool {
		hp++
		h := v.(*handler)
		if h.contentType.Match(ct) {
			go func(h *handler, block interface{}) {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("Recovered while handling", zap.Any("r", r))
					}
				}()
				if err := h.handler(o); err != nil {
					w.logger.Info(
						"Could not handle event",
						zap.String("contentType", ct),
						zap.Error(err),
					)
				}
			}(h, o)
		}
		return true
	})

	if hp == 0 {
		fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		fmt.Println("+++++ NO HANDLERS ++++++++++++++++++++++++++++++++++++++")
		fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		fmt.Println("+++++", ct)
		fmt.Println("+++++", o)
		fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	} else {
		if shouldPersist(ct) {
			bytes, _ := encoding.Marshal(o)
			if err := w.store.Store(o.HashBase58(), bytes); err != nil {
				// TODO handle error
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

	req.response <- payload.RequestedBlock
	return nil
}

func (w *exchange) handleBlockRequest(req *BlockRequest) error {
	blockBytes, err := w.store.Get(req.ID)
	if err != nil {
		return err
	}

	// TODO check if policy allows requested to retrieve the block

	v, err := encoding.Unmarshal(blockBytes)
	if err != nil {
		return err
	}

	resp := &BlockResponse{
		RequestID:      req.RequestID,
		RequestedBlock: v,
	}

	o := resp.ToObject()
	if err := crypto.Sign(o, w.key); err != nil {
		return err
	}

	addr := "peer:" + req.Sender.HashBase58()
	if err := w.Send(context.Background(), o, addr); err != nil {
		w.logger.Warn("blx.handleBlockRequest could not send block", zap.Error(err))
		return err
	}

	return nil
}

func (w *exchange) Get(ctx context.Context, id string) (interface{}, error) {
	// Check local storage for block
	if blockBytes, err := w.store.Get(id); err == nil {
		block, err := encoding.Unmarshal(blockBytes)
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
	// signer := w.addressBook.GetLocalPeerKey()

	// go func() {
	// 	providers, err := w.discovery.GetProviders(ctx, id)
	// 	if err != nil {
	// 		// TODO log err
	// 		return
	// 	}

	// 	for provider := range providers {
	// 		addr := "peer:" + provider.HashBase58()
	// 		if err := w.Send(ctx, req, addr, crypto.SignWith(signer)); err != nil {
	// 			w.logger.Warn("blx.Get could not send req block", zap.Error(err))
	// 		}
	// 	}
	// }()

	for {
		select {
		case payload := <-req.response:
			return payload, nil

		case <-ctx.Done():
			return nil, storage.ErrNotFound
		}
	}
}

func (w *exchange) Send(ctx context.Context, o *encoding.Object, address string) error {
	if shouldPersist(o.GetType()) {
		bytes, _ := encoding.Marshal(o)
		w.store.Store(o.HashBase58(), bytes)
	}

	cerr := make(chan error, 1)
	w.outgoingPayloads <- &outBlock{
		context:   ctx,
		recipient: address,
		block:     o,
		err:       cerr,
	}
	if err := <-cerr; err == nil {
		return nil
	}

	// TODO(geoah) Why is this only handling peers?

	// if recipient is a peer, we might have some relay addresses
	if !strings.HasPrefix(address, "peer:") {
		return ErrAllAddressesFailed
	}

	recipient := strings.Replace(address, "peer:", "", 1)
	if recipient == w.key.GetPublicKey().HashBase58() {
		// TODO(geoah) error or nil?
		return errors.New("cannot send obj to self")
	}

	peer, err := w.net.Resolver().Resolve(recipient)
	if err != nil {
		return err
	}

	if peer == nil {
		w.logger.Warn("wtf")
	}

	// TODO(geoah) make sure fw block is signed
	fw := &BlockForwardRequest{
		Recipient: address,
		FwBlock:   o,
	}

	lk := w.key
	lh := lk.HashBase58()
	fwo := fw.ToObject()
	if err := crypto.Sign(fwo, lk); err != nil {
		return err
	}

	for _, address := range peer.Addresses {
		if strings.HasPrefix(address, "relay:") {
			w.logger.Debug("found relay address", zap.String("peer", recipient), zap.String("address", address))
			relayAddress := strings.Replace(address, "relay:", "", 1)
			// TODO this is an ugly hack
			if strings.Contains(address, lh) {
				continue
			}
			cerr := make(chan error, 1)
			w.outgoingPayloads <- &outBlock{
				context:   ctx,
				recipient: relayAddress,
				block:     fwo,
				err:       cerr,
			}
			if err := <-cerr; err == nil {
				return nil
			}
		}
	}

	return ErrAllAddressesFailed
}

func (w *exchange) GetOrDial(ctx context.Context, address string) (*Connection, error) {
	w.logger.Debug("looking for existing connection", zap.String("address", address))
	if address == "" {
		return nil, errors.New("missing address")
	}

	existingConn, err := w.manager.Get(address)
	if err == nil {
		w.logger.Debug("found existing connection", zap.String("address", address))
		return existingConn, nil
	}

	conn, err := w.net.Dial(ctx, address)
	if err != nil {
		// w.manager.Close(address)
		return nil, errors.Wrap(err, "dial failed")
	}

	go func() {
		if err := w.HandleConnection(conn); err != nil {
			w.logger.Warn("failed to handle block", zap.Error(err))
		}
	}()

	w.manager.Add(address, conn)

	return conn, nil
}

func shouldPersist(t string) bool {
	if strings.Contains(t, "nimona.io/dht") ||
		strings.HasPrefix(t, "nimona.io/telemetry") ||
		strings.HasPrefix(t, "/handshake") {
		return false
	}
	return true
}

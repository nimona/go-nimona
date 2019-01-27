package exchange

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"nimona.io/internal/log"
	nsync "nimona.io/internal/sync"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object"
	"nimona.io/pkg/storage"
)

var (
	// ErrInvalidRequest when received an invalid request object
	ErrInvalidRequest = errors.New("Invalid request")
)

// Exchange interface for mocking exchange
type Exchange interface {
	Get(ctx context.Context, id string) (interface{}, error)
	Handle(contentType string, h func(o *object.Object) error) (func(), error)
	Send(ctx context.Context, o *object.Object, address string) error
}

type exchange struct {
	key     *crypto.Key
	net     net.Network
	manager *ConnectionManager

	outgoing chan *outgoingObject
	incoming chan *incomingObject

	handlers   sync.Map
	logger     *zap.Logger
	streamLock nsync.Kmutex

	store         storage.Storage
	getRequests   sync.Map
	subscriptions sync.Map
}

type outgoingObject struct {
	context   context.Context
	recipient string
	object    *object.Object
	err       chan error
}

type incomingObject struct {
	conn   *net.Connection
	object *object.Object
}

type handler struct {
	contentType glob.Glob
	handler     func(o *object.Object) error
}

// New creates a exchange on a given network
func New(key *crypto.Key, n net.Network, store storage.Storage, address string) (Exchange, error) {
	ctx := context.Background()

	w := &exchange{
		key:     key,
		net:     n,
		manager: &ConnectionManager{},

		outgoing: make(chan *outgoingObject, 10),
		incoming: make(chan *incomingObject, 10),

		handlers:   sync.Map{},
		logger:     log.Logger(ctx).Named("exchange"),
		streamLock: nsync.NewKmutex(),

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
			go func(conn *net.Connection) {
				w.manager.Add("peer:"+conn.RemoteID, conn)
				if err := w.HandleConnection(conn); err != nil {
					w.logger.Warn("failed to handle object", zap.Error(err))
				}
			}(conn)
		}
	}()

	go func() {
		for {
			object := <-w.incoming
			go func(object *incomingObject) {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("Recovered while processing", zap.Any("r", r))
					}
				}()
				if err := w.process(object.object, object.conn); err != nil {
					w.logger.Error("processing object", zap.Error(err))
				}
			}(object)
		}
	}()

	go func() {
		for {
			object := <-w.outgoing
			if object.recipient == "" {
				w.logger.Info("missing recipient")
				object.err <- errors.New("missing recipient")
				continue
			}

			logger := w.logger.With(zap.String("recipient", object.recipient))

			// try to send the object directly to the recipient
			logger.Debug("getting conn to write object")
			conn, err := w.GetOrDial(object.context, object.recipient)
			if err != nil {
				logger.Debug("could not get conn to recipient", zap.Error(err))
				object.err <- err
				continue
			}

			if err := net.Write(object.object, conn); err != nil {
				w.manager.Close(object.recipient)
				logger.Debug("could not write to recipient", zap.Error(err))
				object.err <- err
				continue
			}

			// update peer status
			// TODO(geoah) fix status -- not that we are using this
			// w.addressBook.PutPeerStatus(recipientThumbPrint, peer.StatusConnected)
			object.err <- nil
			continue
		}
	}()

	return w, nil
}

func (w *exchange) Handle(typePatern string, h func(o *object.Object) error) (func(), error) {
	g, err := glob.Compile(typePatern, '.', '/', '#')
	if err != nil {
		return nil, err
	}
	hID := net.RandStringBytesMaskImprSrc(8)
	w.handlers.Store(hID, &handler{
		contentType: g,
		handler:     h,
	})
	r := func() {
		w.handlers.Delete(hID)
	}
	return r, nil
}

func (w *exchange) HandleConnection(conn *net.Connection) error {
	if err := net.Write(w.net.GetPeerInfo().ToObject(), conn); err != nil {
		return err
	}

	w.logger.Debug("handling new connection", zap.String("remote", conn.RemoteID))
	for {
		// TODO use decoder
		object, err := net.Read(conn)
		if err != nil {
			return err
		}

		w.incoming <- &incomingObject{
			conn:   conn,
			object: object,
		}
	}
}

var (
	typeBlockForwardRequest = BlockForwardRequest{}.GetType()
	typeBlockRequest        = BlockRequest{}.GetType()
	typeBlockResponse       = BlockResponse{}.GetType()
)

// Process incoming object
func (w *exchange) process(o *object.Object, conn *net.Connection) error {
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
		w.outgoing <- &outgoingObject{
			context:   ctx,
			recipient: v.Recipient,
			object:    v.FwBlock,
			err:       cerr,
		}
		return <-cerr

	case typeBlockRequest:
		v := &BlockRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		if err := w.handleBlockRequest(v); err != nil {
			w.logger.Warn("could not handle request object", zap.Error(err))
		}

	case typeBlockResponse:
		v := &BlockResponse{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		if err := w.handleBlockResponse(v); err != nil {
			w.logger.Warn("could not handle request object", zap.Error(err))
		}
	}

	ct := o.GetType()

	hp := 0
	w.handlers.Range(func(_, v interface{}) bool {
		hp++
		h := v.(*handler)
		if h.contentType.Match(ct) {
			go func(h *handler, object interface{}) {
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
			bytes, _ := object.Marshal(o)
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
	objectBytes, err := w.store.Get(req.ID)
	if err != nil {
		return err
	}

	// TODO check if policy allows requested to retrieve the object

	v, err := object.Unmarshal(objectBytes)
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
		w.logger.Warn("blx.handleBlockRequest could not send object", zap.Error(err))
		return err
	}

	return nil
}

func (w *exchange) Get(ctx context.Context, id string) (interface{}, error) {
	// Check local storage for object
	if objectBytes, err := w.store.Get(id); err == nil {
		object, err := object.Unmarshal(objectBytes)
		if err != nil {
			return nil, err
		}
		return object, nil
	}

	req := &BlockRequest{
		RequestID: net.RandStringBytesMaskImprSrc(8),
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
	// 			w.logger.Warn("blx.Get could not send req object", zap.Error(err))
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

func (w *exchange) Send(ctx context.Context, o *object.Object, address string) error {
	logger := log.Logger(ctx)

	if shouldPersist(o.GetType()) {
		bytes, _ := object.Marshal(o)
		w.store.Store(o.HashBase58(), bytes)
	}

	switch getAddressType(address) {
	case "peer":
		if err := w.sendDirectlyToPeer(ctx, o, address); err == nil {
			return nil
		}

		if err := w.sendViaRelayToPeer(ctx, o, address); err == nil {
			return nil
		}

		return net.ErrAllAddressesFailed

	case "identity":
		val, err := getAddressValue(address)
		if err != nil {
			return err
		}

		// TODO(geoah): Look for the correct object type as well
		req := &peer.PeerInfoRequest{
			AuthorityKeyHash: val,
		}
		peers, err := w.net.Discoverer().Discover(req)
		if err != nil {
			return errors.New("discovery didn't yield any results for identity")
		}

		failed := 0
		for _, peer := range peers {
			if err := w.sendDirectlyToPeer(ctx, o, peer.Address()); err == nil {
				continue
			}

			if err := w.sendViaRelayToPeer(ctx, o, peer.Address()); err == nil {
				continue
			}

			logger.Info("could not send to peer", zap.String("addr", address))
			failed++
		}

		if failed == len(peers) {
			return errors.New("sending failed for all peers")
		}

		return nil
	}

	return net.ErrAllAddressesFailed
}

func (w *exchange) sendDirectlyToPeer(ctx context.Context, o *object.Object, address string) error {
	cerr := make(chan error, 1)
	w.outgoing <- &outgoingObject{
		context:   ctx,
		recipient: address,
		object:    o,
		err:       cerr,
	}
	if err := <-cerr; err == nil {
		return nil
	}

	return net.ErrAllAddressesFailed
}

func (w *exchange) sendViaRelayToPeer(ctx context.Context, o *object.Object, address string) error {
	// if recipient is a peer, we might have some relay addresses
	if !strings.HasPrefix(address, "peer:") {
		return net.ErrAllAddressesFailed
	}

	recipient := strings.Replace(address, "peer:", "", 1)
	if recipient == w.key.GetPublicKey().HashBase58() {
		// TODO(geoah) error or nil?
		return errors.New("cannot send obj to self")
	}

	q := &peer.PeerInfoRequest{
		SignerKeyHash: recipient,
	}
	peers, err := w.net.Discoverer().Discover(q)
	if err != nil {
		return err
	}

	if len(peers) == 0 {
		return net.ErrNoAddresses
	}

	peer := peers[0]

	// TODO(geoah) make sure fw object is signed
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
			w.outgoing <- &outgoingObject{
				context:   ctx,
				recipient: relayAddress,
				object:    fwo,
				err:       cerr,
			}
			if err := <-cerr; err == nil {
				return nil
			}
		}
	}

	return net.ErrAllAddressesFailed
}
func (w *exchange) GetOrDial(ctx context.Context, address string) (*net.Connection, error) {
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
			w.logger.Warn("failed to handle object", zap.Error(err))
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

func getAddressType(addr string) string {
	return strings.Split(addr, ":")[0]
}

func getAddressValue(addr string) (string, error) {
	ps := strings.Split(addr, ":")
	if len(ps) != 1 {
		return "", errors.New("invalid address")
	}
	return ps[1], nil
}

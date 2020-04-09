package connmanager

import (
	"errors"
	"sync"

	"nimona.io/internal/generator/queue"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/log"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/genny -in=$GENERATORS/syncmap_named/syncmap.go -out=addresses_generated.go -pkg=connmanager gen "KeyType=string ValueType=addressState SyncmapName=addresses"
//go:generate $GOBIN/genny -in=$GENERATORS/syncmap_named/syncmap.go -out=connections_generated.go -imp=nimona.io/pkg/crypto -pkg=connmanager gen "KeyType=crypto.PublicKey ValueType=peerbox SyncmapName=connections"

type addressState int

type peerbox struct {
	peer      crypto.PublicKey
	addresses *AddressesMap
	conn      *net.Connection
	connLock  sync.RWMutex
	queue     *queue.Queue
}

type Manager interface {
	GetConnection(context.Context, *peer.Peer) (*net.Connection, error)
	SetHandler(ConnectionHandler)
}

type ConnectionHandler func(crypto.PublicKey, object.Object)

type manager struct {
	net net.Network

	// store the connections per peer
	connections *ConnectionsMap
	local       *peer.LocalPeer
	connHandler ConnectionHandler
}

func New(
	ctx context.Context,
	n net.Network,
	localInfo *peer.LocalPeer,
) (Manager, error) {
	logger := log.
		FromContext(ctx).
		Named("connmanager").
		With(
			log.String("method", "connmanager.New"),
			log.String("local.peer", localInfo.GetPeerPublicKey().String()),
		)

	mgr := &manager{
		net:         n,
		connections: NewConnectionsMap(),
		local:       localInfo,
	}

	incomingConnections, err := n.Listen(ctx)
	if err != nil {
		return nil, err
	}

	// handle new incoming connections
	go func() {
		for {
			conn := <-incomingConnections
			go func(conn *net.Connection) {
				if err := mgr.handleConnection(conn); err != nil {
					// TODO
					logger.Error("failed to handle connection", log.Error(err))
				}
			}(conn)
		}
	}()

	return mgr, nil
}

func (m *manager) SetHandler(handler ConnectionHandler) {
	m.connHandler = handler
}

func (m *manager) GetConnection(
	ctx context.Context,
	peer *peer.Peer,
) (*net.Connection, error) {

	pbox := m.getPeerbox(peer.PublicKey())

	pbox.connLock.RLock()
	if pbox.conn != nil {
		pbox.connLock.RUnlock()
		return pbox.conn, nil
	}

	pbox.connLock.RUnlock()
	conn, err := m.net.Dial(ctx, peer)
	if err != nil {
		// todo log
		return nil, err
	}

	if err := m.handleConnection(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func (m *manager) handleConnection(
	conn *net.Connection,
) error {
	if conn == nil {
		panic(errors.New("missing connection"))
	}

	// find existing peerbox and update it
	pbox := m.getPeerbox(conn.RemotePeerKey)
	m.updateConnection(pbox, conn)

	if err := net.Write(
		m.local.GetSignedPeer().ToObject(),
		conn,
	); err != nil {
		return err
	}

	go func() {
		for {
			payload, err := net.Read(conn)
			// TODO split errors into connection or payload
			// ie a payload that cannot be unmarshalled or verified
			// should not kill the connection
			if err != nil {
				log.DefaultLogger.Warn(
					"failed to read from connection",
					log.Error(err),
				)
				m.updateConnection(pbox, conn)
				return
			}

			log.DefaultLogger.Debug(
				"reading from connection",
				log.String("payload", payload.GetType()),
			)

			if m.connHandler != nil {
				m.connHandler(conn.RemotePeerKey, *payload)
			}

		}
	}()
	return nil
}

func (m *manager) updateConnection(pbox *peerbox, conn *net.Connection) {
	pbox.connLock.Lock()
	if pbox.conn != nil {
		pbox.conn.Close() // nolint: errcheck
	}
	pbox.conn = conn
	pbox.connLock.Unlock()
}

func (m *manager) getPeerbox(peer crypto.PublicKey) *peerbox {
	pbx := &peerbox{
		peer:      peer,
		addresses: NewAddressesMap(),
	}

	pbx, _ = m.connections.GetOrPut(peer, pbx)

	return pbx
}

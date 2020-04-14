package connmanager

import (
	"sync"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/log"
	"nimona.io/pkg/net"
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
}

type Manager interface {
	GetConnection(context.Context, *peer.Peer) (*net.Connection, error)
}

type ConnectionHandler func(*net.Connection) error

type manager struct {
	net net.Network

	// store the connections per peer
	connections *ConnectionsMap
	local       *peer.LocalPeer
	connHandler ConnectionHandler // TODO: (geoah) should this be a slice?
}

func New(
	ctx context.Context,
	n net.Network,
	localInfo *peer.LocalPeer,
	handler ConnectionHandler,
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
		connHandler: handler,
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
				// find existing peerbox and update it
				pbox := mgr.getPeerbox(conn.RemotePeerKey)
				mgr.updateConnection(pbox, conn)

				if err := mgr.connHandler(conn); err != nil {
					// TODO
					logger.Error("failed to handle connection", log.Error(err))
				}
			}(conn)
		}
	}()

	return mgr, nil
}

func (m *manager) GetConnection(
	ctx context.Context,
	pr *peer.Peer,
) (*net.Connection, error) {
	pbox := m.getPeerbox(pr.PublicKey())

	pbox.connLock.RLock()
	if pbox.conn != nil {
		pbox.connLock.RUnlock()
		return pbox.conn, nil
	}

	pbox.connLock.RUnlock()
	conn, err := m.net.Dial(ctx, pr)
	if err != nil {
		// todo log
		return nil, err
	}

	if err := m.connHandler(conn); err != nil {
		return nil, err
	}

	m.updateConnection(pbox, conn)

	return conn, nil
}

func (m *manager) updateConnection(pbox *peerbox, conn *net.Connection) {
	pbox.connLock.Lock()
	if pbox.conn != nil {
		pbox.conn.Close() // nolint: errcheck
	}
	pbox.conn = conn
	pbox.connLock.Unlock()
}

func (m *manager) getPeerbox(pr crypto.PublicKey) *peerbox {
	pbx := &peerbox{
		peer:      pr,
		addresses: NewAddressesMap(),
	}

	pbx, _ = m.connections.GetOrPut(pr, pbx)

	return pbx
}

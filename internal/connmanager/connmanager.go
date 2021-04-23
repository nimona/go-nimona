package connmanager

import (
	"fmt"
	"sync"

	"nimona.io/internal/net"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

//go:generate genny -in=$GENERATORS/syncmap_named/syncmap.go -out=addresses_generated.go -pkg=connmanager gen "KeyType=string ValueType=addressState SyncmapName=addresses"
//go:generate genny -in=$GENERATORS/syncmap_named/syncmap.go -out=connections_generated.go -imp=nimona.io/pkg/crypto -pkg=connmanager gen "KeyType=string ValueType=peerbox SyncmapName=connections"

type addressState int

type peerbox struct {
	peer      crypto.PublicKey
	addresses *AddressesMap
	conn      *net.Connection
	connLock  sync.RWMutex
}

type Manager interface {
	GetConnection(
		context.Context,
		*peer.ConnectionInfo,
	) (*net.Connection, error)
	CloseConnection(
		*net.Connection,
	)
}

type (
	ConnectionCleanup func()
	ConnectionHandler func(*net.Connection, ConnectionCleanup) error
)

type manager struct {
	net net.Network

	// store the connections per peer
	connections *ConnectionsMap
	connHandler ConnectionHandler // TODO: (geoah) should this be a slice?
}

func New(
	ctx context.Context,
	n net.Network,
	handler ConnectionHandler,
) Manager {
	mgr := &manager{
		net:         n,
		connections: NewConnectionsMap(),
		connHandler: handler,
	}

	// handle new incoming connections
	go func() {
		for {
			conn, _ := n.Accept() // nolint: errcheck
			go func(conn *net.Connection) {
				// find existing peerbox and update it
				pbox := mgr.getPeerbox(conn.RemotePeerKey)
				mgr.updateConnection(pbox, conn)
				// make a cleanup function
				clsr := func() {
					mgr.removeConnection(pbox, conn.ID)
				}
				// and pass the connection and cleanup to the handler
				mgr.connHandler(conn, clsr) // nolint: errcheck
			}(conn)
		}
	}()

	return mgr
}

func (m *manager) GetConnection(
	ctx context.Context,
	pr *peer.ConnectionInfo,
) (*net.Connection, error) {
	pbox := m.getPeerbox(pr.PublicKey)

	pbox.connLock.RLock()
	if pbox.conn != nil {
		conn := pbox.conn
		pbox.connLock.RUnlock()
		return conn, nil
	}

	pbox.connLock.RUnlock()
	if len(pr.Addresses) == 0 {
		return nil, fmt.Errorf("no addresses to dial")
	}

	conn, err := m.net.Dial(ctx, pr)
	if err != nil {
		// todo log
		return nil, err
	}

	m.updateConnection(pbox, conn)

	clsr := func() {
		m.removeConnection(pbox, conn.ID)
	}
	if err := m.connHandler(conn, clsr); err != nil {
		return nil, err
	}

	return conn, nil
}

func (m *manager) CloseConnection(
	conn *net.Connection,
) {
	if conn == nil {
		return
	}
	pbox := m.getPeerbox(conn.RemotePeerKey)
	m.removeConnection(pbox, conn.ID)
}

func (m *manager) removeConnection(pbox *peerbox, connID string) {
	pbox.connLock.Lock()
	defer pbox.connLock.Unlock()
	if pbox.conn == nil {
		return
	}
	if pbox.conn.ID != connID {
		return
	}
	pbox.conn.Close() // nolint: errcheck
	pbox.conn = nil
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

	pbx, _ = m.connections.GetOrPut(pr.String(), pbx)

	return pbx
}

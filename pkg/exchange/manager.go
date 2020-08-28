package exchange

import (
	"sync"

	"nimona.io/pkg/errors"
	"nimona.io/internal/net"
)

type ConnectionManager struct {
	connections sync.Map // key string, value *Connection
}

func (cm *ConnectionManager) Add(address string, conn *net.Connection) {
	cm.connections.Store(address, conn)
}

func (cm *ConnectionManager) Get(remoteID string) (*net.Connection, error) {
	existingConn, ok := cm.connections.Load(remoteID)
	if !ok {
		return nil, errors.New("no stored connection")
	}
	return existingConn.(*net.Connection), nil
}

func (cm *ConnectionManager) Close(fingerprint string) {
	existingConn, ok := cm.connections.Load(fingerprint)
	if !ok {
		return
	}
	existingConn.(*net.Connection).Close() // nolint: errcheck
	cm.connections.Delete(fingerprint)
}

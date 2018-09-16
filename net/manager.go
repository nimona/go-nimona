package net

import (
	"errors"
	"sync"

	"go.uber.org/zap"
	"nimona.io/go/log"
)

type ConnectionManager struct {
	connections sync.Map // key string, value *Connection
}

func (cm *ConnectionManager) Add(conn *Connection) {
	// cm.Close(conn.RemoteID)
	log.DefaultLogger.Warn("_______ STORING", zap.String("remoteID", conn.RemoteID))
	cm.connections.Store(conn.RemoteID, conn)
}

func (cm *ConnectionManager) Get(remoteID string) (*Connection, error) {
	existingConn, ok := cm.connections.Load(remoteID)
	if !ok {
		return nil, errors.New("no stored connection")
	}

	return existingConn.(*Connection), nil
}

func (cm *ConnectionManager) Close(peerID string) {
	log.DefaultLogger.Warn("_____________ CLOSING CONN")
	existingConn, ok := cm.connections.Load(peerID)
	if !ok {
		return
	}
	existingConn.(*Connection).Conn.Close()
	cm.connections.Delete(peerID)
}

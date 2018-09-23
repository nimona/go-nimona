package net

import (
	"errors"
	"sync"
)

type ConnectionManager struct {
	connections sync.Map // key string, value *Connection
}

func (cm *ConnectionManager) Add(conn *Connection) {
	// cm.Close(conn.RemoteID)
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
	existingConn, ok := cm.connections.Load(peerID)
	if !ok {
		return
	}
	existingConn.(*Connection).Conn.Close()
	cm.connections.Delete(peerID)
}

package localpeer

import (
	"sync"

	"nimona.io/internal/rand"
	"nimona.io/pkg/crypto"
)

//go:generate mockgen -destination=../localpeermock/localpeermock_generated.go -package=localpeermock -source=localpeer.go
//go:generate genny -in=$GENERATORS/synclist/synclist.go -out=hashes_generated.go -imp=nimona.io/pkg/chore -pkg=localpeer gen "KeyType=chore.Hash"

type (
	LocalPeer interface {
		GetPeerKey() crypto.PrivateKey
		SetPeerKey(crypto.PrivateKey)
		ListenForUpdates() (<-chan UpdateEvent, func())
	}
	localPeer struct {
		keyLock        sync.RWMutex
		primaryPeerKey crypto.PrivateKey
		listeners      map[string]chan UpdateEvent
		listenersLock  sync.RWMutex
	}
	UpdateEvent string
)

const (
	EventIdentityUpdated UpdateEvent = "identityPublicUpdated"
)

func New() LocalPeer {
	return &localPeer{
		keyLock:       sync.RWMutex{},
		listeners:     map[string]chan UpdateEvent{},
		listenersLock: sync.RWMutex{},
	}
}

func (s *localPeer) SetPeerKey(k crypto.PrivateKey) {
	s.keyLock.Lock()
	s.primaryPeerKey = k
	s.keyLock.Unlock()
}

func (s *localPeer) GetPeerKey() crypto.PrivateKey {
	s.keyLock.RLock()
	defer s.keyLock.RUnlock() //nolint: gocritic
	return s.primaryPeerKey
}

// nolint: unused
func (s *localPeer) publishUpdate(e UpdateEvent) {
	s.listenersLock.RLock()
	defer s.listenersLock.RUnlock()
	for _, l := range s.listeners {
		select {
		case l <- e:
		default:
		}
	}
}

func (s *localPeer) ListenForUpdates() (
	updates <-chan UpdateEvent,
	cancel func(),
) {
	c := make(chan UpdateEvent)
	s.listenersLock.Lock()
	defer s.listenersLock.Unlock()
	id := rand.String(8)
	s.listeners[id] = c
	f := func() {
		s.listenersLock.Lock()
		defer s.listenersLock.Unlock()
		delete(s.listeners, id)
	}
	return c, f
}

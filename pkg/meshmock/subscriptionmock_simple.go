package meshmock

import (
	"sync"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/mesh"
)

type (
	MockSubscriptionSimple struct {
		mutex   sync.Mutex
		index   int
		Objects []*mesh.Envelope
		done    chan struct{}
	}
)

// Cancel the subscription
func (s *MockSubscriptionSimple) Cancel() {
	select {
	case s.done <- struct{}{}:
	default:
	}
}

// Next returns the next object
func (s *MockSubscriptionSimple) Channel() <-chan *mesh.Envelope {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	r := make(chan *mesh.Envelope, len(s.Objects))
	close(r)
	return r
}

// Next returns the next object
func (s *MockSubscriptionSimple) Next() (*mesh.Envelope, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.index >= len(s.Objects) {
		s.done <- struct{}{}
		return nil, errors.Error("done")
	}
	r := s.Objects[s.index]
	s.index++
	return r, nil
}

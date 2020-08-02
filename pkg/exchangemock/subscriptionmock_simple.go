package exchangemock

import (
	"sync"

	"nimona.io/pkg/exchange"
)

type (
	MockSubscriptionSimple struct {
		mutex   sync.Mutex
		index   int
		Objects []*exchange.Envelope
	}
)

// Cancel the subscription
func (s *MockSubscriptionSimple) Cancel() {
}

// Next returns the next object
func (s *MockSubscriptionSimple) Next() (*exchange.Envelope, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.index >= len(s.Objects) {
		return nil, nil
	}
	r := s.Objects[s.index]
	s.index++
	return r, nil
}

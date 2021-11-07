package mesh

import (
	"sync"

	"nimona.io/internal/pubsub"

	"github.com/geoah/go-queue"
)

type (
	mockNextEnvelope struct {
		env *Envelope
		err error
	}
	MockEnvelopeSubscription struct {
		lock  sync.RWMutex
		queue *queue.Queue
	}
)

func (s *MockEnvelopeSubscription) AddNext(env *Envelope, err error) {
	s.lock.Lock()
	if s.queue == nil {
		s.queue = queue.New()
	}
	s.lock.Unlock()
	s.queue.Append(env)
}

func (s *MockEnvelopeSubscription) Next() (*Envelope, error) {
	s.lock.Lock()
	if s.queue == nil {
		s.queue = queue.New()
	}
	s.lock.Unlock()
	v := s.queue.Pop()
	if v == nil {
		return nil, pubsub.ErrSubscriptionCanceled
	}
	return v.(*Envelope), nil
}

func (s *MockEnvelopeSubscription) Cancel() {
}

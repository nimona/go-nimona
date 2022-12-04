package xsync

import (
	"errors"
	"sync"
)

var ErrQueueClosed = errors.New("queue closed")

// Queue is a thread-safe, blocking FIFO queue based on buffered channels
type Queue[T any] struct {
	maxSize int
	items   chan T
	mutex   sync.RWMutex
	closed  bool
}

// NewQueue creates a new Queue with the given maximum size
func NewQueue[T any](maxSize int) *Queue[T] {
	return &Queue[T]{
		maxSize: maxSize,
		items:   make(chan T, maxSize),
		mutex:   sync.RWMutex{},
		closed:  false,
	}
}

// Push adds a new item to the queue
func (q *Queue[T]) Push(item T) error {
	q.mutex.RLock()
	if q.closed {
		q.mutex.RUnlock()
		return ErrQueueClosed
	}
	q.mutex.RUnlock()

	q.items <- item
	return nil
}

// Pop retrieves and removes the next item from the queue
// It blocks until an item is available or the queue is closed
func (q *Queue[T]) Pop() (T, error) {
	item, ok := <-q.items
	if !ok {
		return *new(T), ErrQueueClosed
	}
	return item, nil
}

// Close closes the queue and allows Pop to return
func (q *Queue[T]) Close() {
	q.mutex.Lock()
	q.closed = true
	q.mutex.Unlock()
	close(q.items)
}

// Select returns a channel that can be used in a select statement
func (q *Queue[T]) Select() <-chan T {
	return q.items
}

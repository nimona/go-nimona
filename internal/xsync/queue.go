package xsync

import (
	"fmt"

	"github.com/alphadose/zenq/v2"
)

var ErrQueueClosed = fmt.Errorf("queue is closed")

// Queue is a wrapper around zenq.ZenQ that provides a more tailored API.
type Queue[T any] struct {
	q *zenq.ZenQ[T]
}

// NewQueue creates a new queue.
func NewQueue[T any](size uint32) *Queue[T] {
	return &Queue[T]{
		q: zenq.New[T](size),
	}
}

// Push pushes an item to the queue.
func (q *Queue[T]) Push(item T) error {
	queueClosedForWrites := q.q.Write(item)
	if queueClosedForWrites {
		return ErrQueueClosed
	}
	return nil
}

// Pop pops an item from the queue.
func (q *Queue[T]) Pop() (T, error) {
	item, queueOpen := q.q.Read()
	if !queueOpen {
		return item, ErrQueueClosed
	}
	return item, nil
}

// Close closes the queue.
func (q *Queue[T]) Close() {
	q.q.CloseAsync()
}

// Select returns a channel that can be used to select on.
func (q *Queue[T]) Select() <-chan T {
	c := make(chan T)
	go func() {
		for {
			item := zenq.Select(q.q)
			if item == nil {
				close(c)
				return
			}
			c <- item.(T)
		}
	}()
	return c
}

package queue

import (
	"github.com/cheekybits/genny/generic"

	"nimona.io/internal/context"
	"nimona.io/internal/errors"
)

type (
	ObservableType generic.Type // nolint
	// Queue -
	Queue interface {
		Enqueue(key string, work WorkFunc) error
	}
	item struct {
		key  string
		work WorkFunc
		done chan ObservableType
	}
	q struct {
		cache map[string]bool
		queue chan *item
		done  chan ObservableType
		exit  chan bool
		ctx   context.Context
	}
	// WorkFunc defines work that must be completed
	WorkFunc func() (ObservableType, error)
)

// NewQueue constructs a new queue
func NewQueue(ctx context.Context, workers int, done chan ObservableType) Queue {
	ctx = context.New(ctx, context.WithMethod("queue.New"))
	ctx, cf := context.WithCancel(ctx)

	q := &q{
		cache: map[string]bool{},
		queue: make(chan *item, 10),
		done:  done,
		exit:  make(chan bool),
		ctx:   ctx,
	}

	go func() {
		<-q.exit
		cf()
		close(q.queue)
		close(q.exit)
	}()

	for i := 0; i < workers; i++ {
		go q.process(ctx)
	}

	return q
}

func (q *q) process(ctx context.Context) {
	for {
		select {
		case <-q.ctx.Done():
			return

		case i := <-q.queue:
			if _, ok := q.cache[i.key]; ok {
				continue
			}
			q.cache[i.key] = true
			go func() {
				v, err := i.work()
				if err != nil {
					// TODO log and/or error
				}
				i.done <- v
			}()
		}
	}
}

// Enqueue function with unique key for processing
func (q *q) Enqueue(key string, work WorkFunc) error {
	if key == "" {
		return errors.Error("key is required")
	}
	q.queue <- &item{
		key:  key,
		work: work,
		done: q.done,
	}
	return nil
}

// Stop workers
func (q *q) Stop() {
	q.exit <- true
}

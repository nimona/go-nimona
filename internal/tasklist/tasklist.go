package tasklist

import (
	"sync"

	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
)

const (
	ErrDone = errors.Error("all tasks done")
)

type (
	status   int
	TaskList struct {
		ctx    context.Context
		total  int
		left   int
		queue  chan interface{}
		tasks  map[interface{}]status
		mutex  *sync.Mutex
		done   chan int
		closed bool
	}
)

const (
	StatusIgnored status = iota
	StatusPending
	StatusProcessing
	StatusErrored
	StatusFinished
)

func New(ctx context.Context) *TaskList {
	l := &TaskList{
		ctx:   ctx,
		queue: make(chan interface{}, 100),
		tasks: map[interface{}]status{},
		mutex: &sync.Mutex{},
		done:  make(chan int),
	}
	return l
}

func (l *TaskList) Put(t interface{}) (status, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.closed {
		return StatusIgnored, ErrDone
	}

	s, exists := l.tasks[t]
	if exists {
		return s, nil
	}

	l.total++
	l.left++
	l.tasks[t] = StatusPending
	l.queue <- t
	return StatusPending, nil
}

func (l *TaskList) Ignore(t interface{}) {
	l.mutex.Lock()
	l.tasks[t] = StatusIgnored
	l.mutex.Unlock()
}

func (l *TaskList) Pop() (task interface{}, done func(error), err error) {
	var t interface{}
	select {
	case nt, ok := <-l.queue:
		if !ok {
			return nil, nil, ErrDone
		}
		t = nt
	case <-l.ctx.Done():
		return nil, nil, ErrDone
	}

	l.mutex.Lock()
	l.tasks[t] = StatusProcessing
	l.mutex.Unlock()

	f := func(err error) {
		// acquire a lock
		l.mutex.Lock()
		// mark the task as done
		if err != nil {
			l.tasks[t] = StatusErrored
		} else {
			l.tasks[t] = StatusFinished
		}
		// remove one from left
		l.left--
		// check if the task list is done and mark it as closed
		if l.left == 0 {
			l.closed = true
			// close the queue so any blocked Pop()s return
			close(l.queue)
		}
		// try to push to the done channel
		select {
		case l.done <- l.left:
		default:
		}
		// release lock
		l.mutex.Unlock()
	}

	return t, f, nil
}

func (l *TaskList) Wait() {
	for {
		select {
		case <-l.ctx.Done():
			return
		case l := <-l.done:
			if l == 0 {
				return
			}
		}
	}
}

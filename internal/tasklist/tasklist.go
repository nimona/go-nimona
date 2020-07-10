package tasklist

import (
	"errors"
	"sync"

	"nimona.io/pkg/context"

	"github.com/geoah/go-queue"
)

type (
	status   int
	TaskList struct {
		total int
		left  int
		queue *queue.Queue
		tasks map[interface{}]status
		mutex *sync.Mutex
		done  chan int
	}
)

const (
	StatusPending status = iota
	StatusProcessing
	StatusErrored
	StatusFinished
)

func New() *TaskList {
	l := &TaskList{
		queue: queue.New(),
		tasks: map[interface{}]status{},
		mutex: &sync.Mutex{},
		done:  make(chan int, 1),
	}
	return l
}

func (l *TaskList) Put(t interface{}) status {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	s, exists := l.tasks[t]
	if exists {
		return s
	}

	l.total++
	l.left++
	l.tasks[t] = StatusPending
	l.queue.Append(t)
	return StatusPending
}

func (l *TaskList) Pop() (task interface{}, done func(error), err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.total > 0 && l.left == 0 {
		return nil, nil, errors.New("done")
	}

	t := l.queue.Pop()

	l.tasks[t] = StatusProcessing

	f := func(err error) {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.left--
		select {
		case l.done <- l.left:
		default:
		}
		if err != nil {
			l.tasks[t] = StatusErrored
		} else {
			l.tasks[t] = StatusFinished
		}
	}

	return t, f, nil
}

func (l *TaskList) Wait(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case l := <-l.done:
			if l == 0 {
				return
			}
		}
	}
}

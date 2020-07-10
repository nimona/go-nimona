package tasklist

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"nimona.io/pkg/context"

	"github.com/stretchr/testify/assert"
)

func Test_TaskList_Simple(t *testing.T) {
	l := New()

	// add new task
	s1 := l.Put(1)
	assert.Equal(t, StatusPending, s1)
	assert.Len(t, l.tasks, 1)

	// add one more task
	s2 := l.Put(2)
	assert.Equal(t, StatusPending, s2)
	assert.Len(t, l.tasks, 2)

	// re-add the same task, should not be added
	s2b := l.Put(2)
	assert.Equal(t, StatusPending, s2b)
	assert.Len(t, l.tasks, 2)

	// pop first task
	t1, d1, err := l.Pop()
	assert.NoError(t, err)
	assert.Equal(t, t1, 1)

	// complete first task
	d1(nil)

	// wait for all tasks to be done
	done := int64(0)
	go func() {
		l.Wait(
			context.New(
				context.WithTimeout(time.Second),
			),
		)
		atomic.AddInt64(&done, 1)
	}()

	// task list should not be done yet
	time.Sleep(time.Millisecond)
	assert.Equal(t, int64(0), atomic.LoadInt64(&done))

	// pop second task
	t2, d2, err := l.Pop()
	assert.NoError(t, err)
	assert.Equal(t, t2, 2)

	// pop another task, should block and eventually error
	go func() {
		_, _, err := l.Pop()
		assert.Error(t, err)
	}()

	// complete second task with error
	d2(errors.New("something bad"))

	// task list should now be done
	time.Sleep(time.Millisecond)
	assert.Equal(t, int64(1), atomic.LoadInt64(&done))

	// pop another task, should error with "done"
	_, _, err = l.Pop()
	assert.Error(t, err)
}

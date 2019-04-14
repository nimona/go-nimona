package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/internal/errors"
)

func TestWrap(t *testing.T) {
	var (
		errA = errors.New("a")
		errB = errors.New("b")
		errC = errors.New("c")
	)

	err := errors.Wrap(errB, errA)
	assert.True(t, errors.CausedBy(err, errA))
	assert.True(t, errors.CausedBy(err, errB))
	assert.False(t, errors.CausedBy(err, errC))
	assert.Equal(t, errA, errors.Unwrap(err))
	assert.Equal(t, errB, err.(interface{ Latest() error }).Latest())
	assert.Equal(t, errA, err.(interface{ Cause() error }).Cause())
	assert.Equal(t, "a", errA.Error())
	assert.Equal(t, "b", errB.Error())
	assert.Equal(t, "b: a", err.Error())

	err = errors.Wrap(errB, nil)
	assert.Equal(t, "b", err.Error())
	assert.Nil(t, errors.Unwrap(err))

	err = stderrors.New("d")
	assert.Nil(t, errors.Unwrap(err))

	assert.False(t, errors.CausedBy(nil, errors.New("something")))
}

package backlog

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/asdine/storm"
	"github.com/stretchr/testify/assert"

	"nimona.io/internal/errors"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func TestBolt_PushAndPop(t *testing.T) {
	tf, err := ioutil.TempDir("", "nimona-test-backlog-bolt")
	assert.NoError(t, err)

	st, err := storm.Open(path.Join(tf, "TestBolt_PushAndPop.db"))
	assert.NoError(t, err)
	defer st.Close() // nolint

	bl, err := NewBolt(st)
	assert.NoError(t, err)

	eo1 := object.FromMap(map[string]interface{}{
		"foo:s": "bar",
	})

	eo2 := object.FromMap(map[string]interface{}{
		"foo:s": "bar2",
	})

	k, err := crypto.GenerateKey()
	assert.NoError(t, err)

	// push
	err = bl.Push(eo1, k)
	assert.NoError(t, err)

	// pushing the same obj/key fails
	err = bl.Push(eo1, k)
	assert.Error(t, err)
	assert.True(t, errors.CausedBy(err, ErrAlreadyExists))

	// push one more just to check order
	err = bl.Push(eo2, k)
	assert.NoError(t, err)

	// pop
	ao, _, err := bl.Pop(k)
	assert.NoError(t, err)
	assert.Equal(t, eo1.HashBase58(), ao.HashBase58())

	// pop the second one
	ao, _, err = bl.Pop(k)
	assert.NoError(t, err)
	assert.Equal(t, eo2.HashBase58(), ao.HashBase58())

	// pop should error
	ao, _, err = bl.Pop(k)
	assert.Error(t, err)
	assert.True(t, errors.CausedBy(err, ErrNoMoreObjects))
	assert.Nil(t, ao)
}

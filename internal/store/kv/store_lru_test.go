package kv

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StoreLRU_HappyPathSuccess(t *testing.T) {
	s := NewLRU(3)

	value := []byte("bar")
	key := "foo"

	err := s.Put(key, value)
	assert.NoError(t, err)

	v, err := s.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, value, v)

	err = s.Put(key, value)
	assert.NoError(t, err)

	l, err := s.List()
	assert.NoError(t, err)
	assert.Len(t, l, 1)

	err = s.Put("foo.bar", []byte{1, 2, 3})
	assert.NoError(t, err)

	err = s.Put("foo.foo.bar", []byte{1, 2, 3})
	assert.NoError(t, err)

	v, err = s.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, value, v)

	err = s.Check(key)
	assert.NoError(t, err)

	err = s.Put("foo.bar.second", []byte{1, 2, 3})
	assert.NoError(t, err)

	l, err = s.List()
	assert.NoError(t, err)
	assert.Len(t, l, 3)

	v, err = s.Get("foo.bar")
	assert.Error(t, err)
	assert.Nil(t, v)

	err = s.Check("foo.bar")
	assert.Error(t, err)

	list, err := s.Scan("foo")
	sort.Strings(list)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "foo.bar.second", "foo.foo.bar"}, list)

	list, err = s.Scan("foo.bar")
	sort.Strings(list)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo.bar.second"}, list)
}

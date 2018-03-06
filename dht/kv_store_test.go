package dht

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

func TestFindKeysNearestTo(t *testing.T) {
	s, err := newStore()
	assert.Nil(t, err)

	s.Put(KeyPrefixPeer+"a1", "0.0.0.0", true)
	s.Put(KeyPrefixPeer+"a2", "0.0.0.1", true)
	s.Put(KeyPrefixPeer+"a3", "0.0.0.3", true)
	s.Put(KeyPrefixPeer+"a4", "0.0.0.4", true)
	s.Put(KeyPrefixPeer+"a5", "0.0.0.5", true)

	k1, err := s.FindKeysNearestTo(KeyPrefixPeer, KeyPrefixPeer+"a1", 1)
	assert.Nil(t, err)

	k2, err := s.FindKeysNearestTo(KeyPrefixPeer, KeyPrefixPeer+"a2", 1)
	assert.Nil(t, err)

	assert.NotEqual(t, k1[0], k2[0])
	assert.Equal(t, trimKey(k1[0], KeyPrefixPeer), "a1")
	assert.Equal(t, trimKey(k2[0], KeyPrefixPeer), "a2")

}

func TestPut(t *testing.T) {
	s, err := newStore()
	assert.Nil(t, err)

	key := "k1"
	value := "v1"
	persistent := false

	err = s.Put(key, value, persistent)
	assert.Nil(t, err)

	assert.NotEmpty(t, s.pairs[key])
	assert.Equal(t, s.pairs[key][0].Value, value)
	assert.Equal(t, s.pairs[key][0].Persistent, persistent)
}

func TestGet(t *testing.T) {
	s, err := newStore()
	assert.Nil(t, err)

	key := "k1"
	value := "v1"
	persistent := false

	p := &Pair{
		Key:        key,
		Value:      value,
		Persistent: persistent,
		LastPut:    time.Now(),
	}
	s.pairs[key] = append(s.pairs[key], p)

	pairs, err := s.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, pairs[0], value)
}

func TestWipe(t *testing.T) {
	s, err := newStore()
	assert.Nil(t, err)

	key := "k1"
	value := "v1"
	persistent := false

	p := &Pair{
		Key:        key,
		Value:      value,
		Persistent: persistent,
		LastPut:    time.Now(),
	}
	s.pairs[key] = append(s.pairs[key], p)

	err = s.Wipe(key)
	assert.Nil(t, err)
	assert.Empty(t, s.pairs)
}

func TestGetAll(t *testing.T) {
	s, err := newStore()
	assert.Nil(t, err)

	key1 := "k1"
	value1 := "v1"
	persistent1 := false
	p1 := &Pair{
		Key:        key1,
		Value:      value1,
		Persistent: persistent1,
		LastPut:    time.Now(),
	}

	key2 := "k2"
	value2 := "v2"
	persistent2 := false
	p2 := &Pair{
		Key:        key2,
		Value:      value2,
		Persistent: persistent2,
		LastPut:    time.Now(),
	}

	s.pairs[key1] = append(s.pairs[key1], p1)
	s.pairs[key2] = append(s.pairs[key2], p2)

	allPairs, err := s.GetAll()
	assert.Nil(t, err)
	assert.Len(t, allPairs, 2)
}

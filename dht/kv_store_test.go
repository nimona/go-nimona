package dht

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestFindKeysNearestToNotEqual(t *testing.T) {
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

	assert.NotEqual(t, k1, k2)
}

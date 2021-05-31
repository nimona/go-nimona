package object

// import (
// 	"testing"

// 	"nimona.io/pkg/crypto"

// 	"github.com/stretchr/testify/assert"
// )

// func TestNewSignature(t *testing.T) {
// 	sk, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
// 	assert.NoError(t, err)

// 	o := FromMap(Map{
// 		"foo:s": String("bar"),
// 	})

// 	sig, err := NewSignature(sk, o)
// 	assert.NoError(t, err)
// 	assert.Equal(t, sk.PublicKey(), sig.Signer)

// 	o.Metadata.Signature = sig

// 	err = Verify(o)
// 	assert.NoError(t, err)
// }

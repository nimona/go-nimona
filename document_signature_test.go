package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentSignature_SignVerify(t *testing.T) {
	_, sk, err := GenerateKey()
	require.NoError(t, err)

	doc := &CborFixture{
		String: "foo",
	}

	hash, err := NewDocumentHash(doc)
	require.NoError(t, err)

	sig, err := NewDocumentSignature(sk, hash)
	require.NoError(t, err)

	err = VerifySignature(sig, hash)
	require.NoError(t, err)
}

package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestRequestContext(t *testing.T) *RequestContext {
	t.Helper()

	pk, sk, err := GenerateKey()
	require.NoError(t, err)

	return &RequestContext{
		PrivateKey: sk,
		PublicKey:  pk,
	}
}

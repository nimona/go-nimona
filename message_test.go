package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageWrapper(t *testing.T) {
	type TestBody struct {
		Test string `cbor:"test"`
	}
	m0 := &MessageWrapper[TestBody]{
		Type: "test",
		Body: TestBody{
			Test: "test-body",
		},
	}
	m0Bytes, err := m0.MarshalCBOR()
	require.NoError(t, err)

	m1 := &MessageWrapper[TestBody]{}
	err = m1.UnmarshalCBOR(m0Bytes)
	require.NoError(t, err)

	require.Equal(t, m0, m1)
}

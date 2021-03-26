package time

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDateTime(t *testing.T) {
	now := Now()
	nowString, err := now.MarshalString()
	require.NoError(t, err)
	require.NotEmpty(t, nowString)

	got := DateTime{}
	err = got.UnmarshalString(nowString)
	require.NoError(t, err)
	require.Equal(t, now.Unix(), got.Unix())

	err = got.UnmarshalString("not a date")
	require.Error(t, err)
}

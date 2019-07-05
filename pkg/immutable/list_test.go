package immutable

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	l := List{}
	require.Equal(t, 0, l.Length())

	iCalls := 0
	l.Iterate(func(_ Value) {
		iCalls++
	})
	require.Equal(t, 0, iCalls)

	l = l.Append(Value{stringValue{"foo"}})
	require.Equal(t, 1, l.Length())

	l1 := l.Append(Value{stringValue{"bar"}})
	require.Equal(t, 2, l1.Length())

	l2 := l.Append(Value{stringValue{"bar2"}})
	require.Equal(t, 2, l2.Length())

	iCalls = 0
	values := []string{}
	l1.Iterate(func(v Value) {
		iCalls++
		values = append(values, v.StringValue())
	})
	require.Equal(t, 2, iCalls)
	require.Len(t, values, 2)
	require.Equal(t, []string{
		"foo",
		"bar",
	}, values)
}

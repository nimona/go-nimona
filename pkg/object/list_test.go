package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	l := List{}
	require.Equal(t, 0, l.Length())

	iCalls := 0
	l.Iterate(func(_ int, _ Value) bool {
		iCalls++
		return true
	})
	require.Equal(t, 0, iCalls)

	l = l.Append(String("foo"))
	require.Equal(t, 1, l.Length())

	l1 := l.Append(String("bar"))
	require.Equal(t, 2, l1.Length())

	l2 := l.Append(String("bar2"))
	require.Equal(t, 2, l2.Length())

	iCalls = 0
	values := []string{}
	l1.Iterate(func(_ int, v Value) bool {
		iCalls++
		values = append(values, v.PrimitiveHinted().(string))
		return true
	})
	require.Equal(t, 2, iCalls)
	require.Len(t, values, 2)
	require.Equal(t, []string{"foo", "bar"}, values)

	assert.Equal(t, []string{"foo", "bar2"}, l2.PrimitiveHinted())

	l2a := l2.Set("0", String("not-foo"))
	assert.Equal(t, []string{"not-foo", "bar2"}, l2a.PrimitiveHinted())

	l2b := l2a.Set("1", String("not-foo"))
	assert.Equal(t, []string{"not-foo", "not-foo"}, l2b.PrimitiveHinted())

	la := List{}
	la = la.Set("10", String("10"))
	assert.Equal(t, []string{"10"}, la.PrimitiveHinted())

	la = la.Set("1", String("1"))
	assert.Equal(t, []string{"1", "10"}, la.PrimitiveHinted())

	la = la.Append(String("11"))
	assert.Equal(t, []string{"1", "10", "11"}, la.PrimitiveHinted())

	la = la.Set("9", String("9"))
	assert.Equal(t, []string{"1", "9", "10", "11"}, la.PrimitiveHinted())

	la = la.Set("0", String("0"))
	assert.Equal(t, []string{"0", "1", "9", "10", "11"}, la.PrimitiveHinted())
}

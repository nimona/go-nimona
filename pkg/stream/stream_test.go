package stream

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nimona.io/pkg/object"
)

type (
	TestState struct {
		Foo      int
		Bar      int
		Comments []string
	}
	TestEventAddFoo struct {
		Metadata object.Metadata `nimona:"@metadata:m,type=test/foobar.AddFoo"`
	}
	TestEventAddBar struct {
		Metadata object.Metadata `nimona:"@metadata:m,type=test/foobar.AddBar"`
	}
	TestEventAddComment struct {
		Metadata object.Metadata `nimona:"@metadata:m,type=test/foobar.Comment"`
		Comment  string          `nimona:"comment:s"`
	}
)

func (e *TestEventAddFoo) Apply(s *TestState) error {
	s.Foo++
	return nil
}

func (e *TestEventAddBar) Apply(s *TestState) error {
	s.Bar++
	return nil
}

func (e *TestEventAddComment) Apply(s *TestState) error {
	s.Comments = append(s.Comments, e.Comment)
	return nil
}

func Test_Controller_Example(t *testing.T) {
	c, err := NewController[TestState](
		&TestEventAddFoo{},
		&TestEventAddBar{},
		&TestEventAddComment{},
	)
	require.NoError(t, err)
	require.NotNil(t, c)

	s := c.NewStream()
	err = s.ApplyEvent(&TestEventAddFoo{})
	require.NoError(t, err)
	err = s.ApplyEvent(&TestEventAddFoo{})
	require.NoError(t, err)
	err = s.ApplyEvent(&TestEventAddBar{})
	require.NoError(t, err)
	err = s.ApplyEvent(&TestEventAddComment{Comment: "hello"})
	require.NoError(t, err)
	err = s.ApplyEvent(&TestEventAddComment{Comment: "world"})
	require.NoError(t, err)

	require.Equal(t, TestState{
		Foo: 2,
		Bar: 1,
		Comments: []string{
			"hello",
			"world",
		},
	}, s.GetState())
}

package fixtures

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/object"
)

func TestFixtures_Composite(t *testing.T) {
	c := &CompositeTest{
		CompositeStringTest:         &Composite{"foo"},
		CompositeDataTest:           &Composite{"foo"},
		RepeatedCompositeStringTest: []*Composite{{"foo"}, {"bar"}},
		RepeatedCompositeDataTest:   []*Composite{{"foo"}, {"bar"}},
	}

	m, err := c.MarshalMap()
	require.NoError(t, err)

	b, err := json.Marshal(m)
	require.NoError(t, err)

	fmt.Println(string(b))

	gm := &object.Object{}
	err = json.Unmarshal(b, gm)
	require.NoError(t, err)

	gc := &CompositeTest{}
	err = gc.FromObject(gm)
	require.NoError(t, err)
}

func TestFixtures_Nested(t *testing.T) {
	c := &Parent{
		Foo: "bar0",
		Child: &Child{
			Metadata: object.Metadata{
				Stream: "random",
			},
			Foo: "bar1",
		},
		RepeatedChild: []*Child{{
			Metadata: object.Metadata{
				Stream: "random",
			},
			Foo: "bar2",
		}, {
			Metadata: object.Metadata{
				Stream: "random",
			},
			Foo: "bar3",
		}},
	}

	m, err := c.MarshalMap()
	require.NoError(t, err)

	b, err := json.Marshal(m)
	require.NoError(t, err)

	fmt.Println(string(b))

	gm := &object.Object{}
	err = json.Unmarshal(b, gm)
	require.NoError(t, err)

	gc := &Parent{}
	err = gc.FromObject(gm)
	require.NoError(t, err)
}

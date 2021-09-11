package schema

import (
	"encoding/json"
	"fmt"
	"testing"

	"nimona.io/pkg/object"
	"nimona.io/pkg/tilde"

	"github.com/stretchr/testify/require"
)

func TestSchema_Self(t *testing.T) {
	c := &Context{
		Name: "schema/Context",
		Types: []*Type{{
			Properties: []*Property{{
				Name:     "name",
				Hint:     tilde.StringHint,
				Required: true,
			}, {
				Name: "description",
				Hint: tilde.StringHint,
			}, {
				Name: "version",
				Hint: tilde.StringHint,
			}, {
				Name:     "types",
				Hint:     tilde.MapHint,
				Repeated: true,
			}},
		}, {
			Properties: []*Property{{
				Name:     "name",
				Hint:     tilde.StringHint,
				Required: true,
			}, {
				Name: "description",
				Hint: tilde.StringHint,
			}, {
				Name:     "properties",
				Hint:     tilde.MapHint,
				Repeated: true,
			}, {
				Name: "strict",
				Hint: tilde.BoolHint,
			}},
		}, {
			Properties: []*Property{{
				Name:     "name",
				Hint:     tilde.StringHint,
				Required: true,
			}, {
				Name: "description",
				Hint: tilde.StringHint,
			}, {
				Name: "hint",
				Hint: tilde.StringHint,
			}, {
				Name: "type",
				Hint: tilde.MapHint,
			}, {
				Name: "context",
				Hint: tilde.MapHint,
			}, {
				Name: "required",
				Hint: tilde.BoolHint,
			}, {
				Name: "repeated",
				Hint: tilde.BoolHint,
			}},
		}},
	}

	o, err := object.Marshal(c)
	require.NoError(t, err)

	b, err := json.MarshalIndent(o, "", "  ")
	require.NoError(t, err)
	fmt.Println(string(b))
}

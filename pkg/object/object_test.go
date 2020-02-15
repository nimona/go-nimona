package object

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/immutable"
)

func TestObject(t *testing.T) {
	o := Object{
		Header: Header{
			Stream:  Hash("stream"),
			Parents: []Hash{"parent1", "parent2"},
			Policy: Policy{
				Subjects: []string{"subject1", "subject2"},
				Actions:  []string{"action1", "action2"},
				Effect:   "effect",
			},
			Signature: Signature{
				Signer: "signer",
				Alg:    "alg",
				X:      []byte{1, 2, 3},
			},
			Owners: []crypto.PublicKey{"owner1", "owner2"},
		},
		Data: immutable.AnyToValue(":o", map[string]interface{}{
			"foo:s": "bar",
		}).(immutable.Map),
	}
	o.SetType("type")

	m := o.ToMap()
	n := FromMap(m)

	require.EqualValues(t, o.Header, n.Header)
	require.EqualValues(t, o.Data.PrimitiveHinted(), n.Data.PrimitiveHinted())

	jb, err := json.Marshal(m)
	require.NoError(t, err)

	jm := map[string]interface{}{}
	err = json.Unmarshal(jb, &jm)
	require.NoError(t, err)
	n = FromMap(jm)

	require.EqualValues(t, o.Header, n.Header)
	require.EqualValues(t, o.Data.PrimitiveHinted(), n.Data.PrimitiveHinted())
}

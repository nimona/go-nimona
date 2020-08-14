package object

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
)

func TestObject(t *testing.T) {
	o := Object{}
	o = o.SetType("type")
	o = o.SetStream(Hash("stream"))
	o = o.SetParents([]Hash{"parent1", "parent2"})
	o = o.SetPolicy(Policy{
		Subjects: []string{"subject1", "subject2"},
		Actions:  []string{"action1", "action2"},
		Effect:   "effect",
	})
	o = o.SetSignature(Signature{
		Signer: "signer",
		Alg:    "alg",
		X:      []byte{1, 2, 3},
	})
	o = o.SetOwner(crypto.PublicKey("owner1"))
	o = o.Set("foo:s", "bar")

	m := map[string]interface{}{
		"type:s": "type",
		"metadata:m": map[string]interface{}{
			"stream:s":   "stream",
			"parents:as": []string{"parent1", "parent2"},
			"policy:m": map[string]interface{}{
				"subjects:as": []string{"subject1", "subject2"},
				"actions:as":  []string{"action1", "action2"},
				"effect:s":    "effect",
			},
			"_signature:m": map[string]interface{}{
				"signer:s": "signer",
				"alg:s":    "alg",
				"x:d":      []byte{1, 2, 3},
			},
			"owner:s": "owner1",
		},
		"data:m": map[string]interface{}{
			"foo:s": "bar",
		},
	}

	n := FromMap(m)
	require.EqualValues(
		t,
		o.Raw().PrimitiveHinted(),
		n.Raw().PrimitiveHinted(),
	)

	require.EqualValues(t, "type", o.GetType())
	require.EqualValues(t, Hash("stream"), o.GetStream())
	require.EqualValues(t, []Hash{"parent1", "parent2"}, o.GetParents())

	jb, err := json.Marshal(m)
	require.NoError(t, err)

	jm := map[string]interface{}{}
	err = json.Unmarshal(jb, &jm)
	require.NoError(t, err)

	n = FromMap(jm)
	require.EqualValues(t, o.Raw().PrimitiveHinted(), n.Raw().PrimitiveHinted())
}

package object_test

// import (
// 	"encoding/json"
// 	"testing"

// 	"github.com/google/go-cmp/cmp"
// 	"github.com/stretchr/testify/require"

// 	"nimona.io/internal/fixtures"
// 	"nimona.io/pkg/crypto"
// 	"nimona.io/pkg/object"
// )

// func Test_Normalize(t *testing.T) {
// 	kx := []byte{111, 39, 65, 188, 215, 49, 84, 43, 192, 187}

// 	s := fixtures.TestStream{
// 		Header: object.Header{
// 			Signature: object.Signature{
// 				Signer: "foo",
// 				Alg:    "alg",
// 				X:      kx,
// 			},
// 			Owners: []crypto.PublicKey{
// 				"foo",
// 			},
// 			Policy: object.Policy{
// 				Subjects:  []string{"subject"},
// 				Resources: []string{"*"},
// 				Actions:   []string{"allow"},
// 			},
// 		},
// 		Nonce: "nonce",
// 	}

// 	em := map[string]interface{}{
// 		"header:m": map[string]interface{}{
// 			"type:s":    "nimona.io/fixtures.TestStream",
// 			"owners:as": []string{"foo"},
// 			"policy:m": map[string]interface{}{
// 				"subjects:as":  []string{"subject"},
// 				"resources:as": []string{"*"},
// 				"actions:as":   []string{"allow"},
// 			},
// 			"_signature:m": map[string]interface{}{
// 				"signer:s": "foo",
// 				"alg:s":    "alg",
// 				"x:d":      kx,
// 			},
// 		},
// 		"content:m": map[string]interface{}{
// 			"nonce:s": "nonce",
// 		},
// 	}

// 	b, err := json.MarshalIndent(s.ToObject().ToMap(), "", "  ")
// 	require.NoError(t, err)

// 	m := map[string]interface{}{}
// 	require.NoError(t, json.Unmarshal(b, &m))

// 	nm, err := object.Normalize(m)
// 	require.NoError(t, err)

// 	if !cmp.Equal(nm, em) {
// 		t.Errorf("Normalize() result doesn't match expectd " + cmp.Diff(nm, em))
// 	}
// }

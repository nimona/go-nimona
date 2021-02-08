package object

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/encoding/base58"
)

func BenchmarkHash(b *testing.B) {
	o := Map{
		"type": String("blob"),
		"data": Map{
			"filename": String("foo"),
			"dummy": Map{
				"type": String("dummy"),
				"metadata": Map{
					"owner": String("foo"),
				},
				"data": Map{
					"foo": String("bar"),
					"data": Data(
						"1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14" +
							"\n15\n16\n17\n18\n19\n20",
					),
				},
			},
		},
	}
	for n := 0; n < b.N; n++ {
		fromValue(o) // nolint: errcheck
	}
}

func TestFromValue(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    string
		wantErr bool
	}{{
		name: "5",
		json: `{"something:s":"foo","metadata:m":{}}`,
		want: "uvqAvGish5DVzVsZcn9aFvvFG8JgeuwgGMnLj2dfp6e",
	}, {
		name: "6",
		json: `{"something:s":"foo"}`,
		want: "uvqAvGish5DVzVsZcn9aFvvFG8JgeuwgGMnLj2dfp6e",
	}, {
		name: "7",
		json: `{"something:d":"Zm9v"}`,
		want: "GaRxrpcBkBP16rmL4kNzhwZvqo6DjqW5bccNm65okSYG",
	}, {
		name: "8",
		json: `{"something:b":false}`,
		want: "AVigrbFWTBNVXeB5Q6GDrt75sTBrpwuEqMVP7DVfW9mM",
	}, {
		name: "9",
		json: `{"something:b":true}`,
		want: "69CbvTybbM2DPrCRqoGyt7kxFUKhYwSbtUdtCs9HQLve",
	}, {
		name: "10",
		json: `{"something:i":1234567890}`,
		want: "Am2CNoZisskHDL2E8srhPHc4L5wCGUv6nuJjVT6Ca1iV",
	}, {
		name: "11",
		json: `{"something:f":12345.6789}`,
		want: "AnVQHPHdbE5VDo2XG21VRi6yESZWnwKpf3SU68WAspUC",
	}, {
		name: "13",
		json: `{"something:as":["foo","bar"]}`,
		want: "EYCgWPkfYeGew331WYBaKphmtxDgPcJet6pWpSokK9Am",
	}, {
		name: "14",
		json: `{"something:ai":[123,456]}`,
		want: "FzULygYLCUkuEibqPKYxnoEMSnnfaNfkbuJDeZgnZra5",
	}, {
		name: "15",
		json: `{"foo:s":"bar"}`,
		want: "FwXyoLg3qpzM8R8uZECrymsyGuKVTyTn3qoNsmGhEMRg",
	}, {
		name: "17",
		// nolint: lll
		json: `{"data:m":{"foo:s":"bar","nested:m":{"_sig:s":"should not matter","foo:s":"bar"}}}`,
		want: "bmRkoyP1pWmRphQVpCGKJz7EJDY7mEpLNrPW4zRedkj",
	}, {
		name: "18",
		// nolint: lll
		json: `{"data:m":{"foo:s":"bar","nested:m":{"_signature:m":{"foo:s":"bar"},"foo:s":"bar"}}}`,
		want: "bmRkoyP1pWmRphQVpCGKJz7EJDY7mEpLNrPW4zRedkj",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Map{}
			assert.NoError(t, json.Unmarshal([]byte(tt.json), &m))
			got, err := fromValue(m)
			require.NoError(t, err)
			assert.Equal(t, tt.want, base58.Encode(got[:]))
		})
	}
}

func TestRaw(t *testing.T) {
	r := rawHash{12, 13, 14, 15, 15}
	h := hashFromRaw(r)
	g, err := hashToRaw(h)
	require.NoError(t, err)
	require.Equal(t, g, r)
}

func TestNewhash(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Hash
		wantErr bool
	}{{
		name: "1",
		json: `{"type:s":"foo","data:m":{"foo:s":"bar"}}`,
		want: "oh1.D5ZyytQVJ8hLyLHL8PbGyrGkTuYNNzZanHnYATKX1ctN",
	}, {
		name: "2",
		// nolint: lll
		json: `{"type:s":"foo","data:m":{"foo:o":{"type:s":"foo","data:m":{"foo:s":"bar"}}}}`,
		want: "oh1.CCY333XK4N91Fwuunj3N1RGizqPo96JkictfjqHK68XW",
	}, {
		name: "3",
		// nolint: lll
		json: `{"type:s":"foo","data:m":{"foo:h":"oh1.D5ZyytQVJ8hLyLHL8PbGyrGkTuYNNzZanHnYATKX1ctN"}}`,
		want: "oh1.CCY333XK4N91Fwuunj3N1RGizqPo96JkictfjqHK68XW",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Object{}
			assert.NoError(t, json.Unmarshal([]byte(tt.json), o))
			got, err := NewHash(o)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

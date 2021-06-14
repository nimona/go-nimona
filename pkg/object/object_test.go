package object

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/object/cid"
)

func TestObject(t *testing.T) {
	o := chore.Map{
		"boolArray": chore.BoolArray{
			chore.Bool(false),
			chore.Bool(true),
		},
		"dataArray": chore.DataArray{
			chore.Data("v0"),
			chore.Data("v1"),
		},
		"floatArray": chore.FloatArray{
			chore.Float(0.10),
			chore.Float(1.12),
		},
		"intArray": chore.IntArray{
			chore.Int(0),
			chore.Int(1),
		},
		"mapArray": chore.MapArray{
			chore.Map{"foo0": chore.String("bar0")},
			chore.Map{
				"boolArray": chore.BoolArray{
					chore.Bool(false),
					chore.Bool(true),
				},
				"dataArray": chore.DataArray{
					chore.Data("v0"),
					chore.Data("v1"),
				},
				"floatArray": chore.FloatArray{
					chore.Float(0.10),
					chore.Float(1.12),
				},
				"intArray": chore.IntArray{
					chore.Int(0),
					chore.Int(1),
				},
				"mapArray": chore.MapArray{
					chore.Map{"foo0": chore.String("bar0")},
					chore.Map{"foo1": chore.String("bar1")},
				},
				"stringArray": chore.StringArray{
					chore.String("v0"),
					chore.String("v1"),
				},
				"uintArray": chore.UintArray{
					chore.Uint(0),
					chore.Uint(1),
				},
			},
		},
		"stringArray": chore.StringArray{
			chore.String("v0"),
			chore.String("v1"),
		},
		"uintArray": chore.UintArray{
			chore.Uint(0),
			chore.Uint(1),
		},
		"bool":  chore.Bool(true),
		"data":  chore.Data("foo"),
		"float": chore.Float(1.1),
		"int":   chore.Int(2),
		"map": chore.Map{
			"boolArray": chore.BoolArray{
				chore.Bool(false),
				chore.Bool(true),
			},
			"dataArray": chore.DataArray{
				chore.Data("v0"),
				chore.Data("v1"),
			},
			"floatArray": chore.FloatArray{
				chore.Float(0.10),
				chore.Float(1.12),
			},
			"chore.IntArray": chore.IntArray{
				chore.Int(0),
				chore.Int(1),
			},
			"mapArray": chore.MapArray{
				chore.Map{"foo0": chore.String("bar0")},
				chore.Map{"foo1": chore.String("bar1")},
			},
			"stringArray": chore.StringArray{
				chore.String("v0"),
				chore.String("v1"),
			},
			"chore.UintArray": chore.UintArray{
				chore.Uint(0),
				chore.Uint(1),
			},
			"bool":  chore.Bool(true),
			"data":  chore.Data("foo"),
			"float": chore.Float(1.1),
			"int":   chore.Int(2),
			"map": chore.Map{
				"int": chore.Int(42),
			},
			"string": chore.String("foo"),
			"uint":   chore.Uint(3),
		},
		"string": chore.String("foo"),
		"uint":   chore.Uint(3),
	}
	b, err := json.MarshalIndent(o, "", "  ")
	require.NoError(t, err)

	g := chore.Map{}
	err = json.Unmarshal(b, &g)
	require.NoError(t, err)

	require.Equal(t, o, g)
}

// nolint: lll
func TestValues(t *testing.T) {
	dummyObj := &Object{
		Type: "dummy",
		Data: chore.Map{
			"foo": chore.String("bar"),
		},
	}
	dummy, err := dummyObj.MarshalMap()
	require.NoError(t, err)
	dummyCID, err := cid.New(dummy)
	require.NoError(t, err)
	tests := []struct {
		name    string
		value   chore.Value
		json    string
		want    string
		wantErr bool
	}{{
		name:  "s",
		value: chore.String("foo"),
		want:  "Qmczn3TuD5jpokERZczuuPXbgJikFgJkwgr6J9p3dnTSkj",
		json:  `"foo"`,
	}, {
		name:  "d",
		value: chore.Data([]byte{0, 1, 2}),
		want:  "QmXZgMaXSSJsRg9zstbj6nURqGVppGMBKxGSUrNU8u7D2r",
		json:  `"AAEC"`,
	}, {
		name:  "b0",
		value: chore.Bool(false),
		want:  "QmQP6XKQetJQ1ZQNm5SoC7skxj28J4QAGhSnrHkEA2gNWU",
		json:  `false`,
	}, {
		name:  "b1",
		value: chore.Bool(true),
		want:  "QmPR2X9b88aGKUTEh7UuFdcvG81RQ72jXUsMHob1qRZFwa",
		json:  `true`,
	}, {
		name:  "i",
		value: chore.Int(1234567890),
		want:  "QmcaJDHHUUsiwcYk554TCLMQxvDYvgdH9f1n4MAM9JuqC9",
		json:  `1234567890`,
	}, {
		name:  "f",
		value: chore.Float(12345.6789),
		want:  "QmdMKevPGKgWFGJCfmWxmMWQpztKHRo9kmF6dyx6MWuSmZ",
		json:  `12345.6789`,
	}, {
		name:  "u",
		value: chore.Uint(123456789),
		want:  "QmanqLLrcWahzVH4VUKsfq44VegJSS6G1gYk5cLUSCYK9x",
		json:  `123456789`,
	}, {
		name:  "as",
		value: chore.StringArray{"foo", "foo"},
		want:  "QmW2K4k56Fxvyk7v6YQ4iEeKtNeCd2m2Nu1xwPQtSR6fho",
		json:  `["foo","foo"]`,
	}, {
		name:  "ad",
		value: chore.DataArray{[]byte{0, 1, 2}, []byte{0, 1, 2}},
		want:  "QmeBcPzgGU1YGdSY5SCr4CVMnA2bopJ4tUN5McaPZVkyi1",
		json:  `["AAEC","AAEC"]`,
	}, {
		name:  "ab0",
		value: chore.BoolArray{false, false},
		want:  "QmcC2yFC2FvrZVpgahacSmBSx34sPcY2Tnh3VSfGgsZjhH",
		json:  `[false,false]`,
	}, {
		name:  "ab1",
		value: chore.BoolArray{true, true},
		want:  "QmXqhrACfxuhKfKoCQegJwrMixxmut84kJcJpiZ8tiSXYi",
		json:  `[true,true]`,
	}, {
		name:  "ai",
		value: chore.IntArray{1234567890, 1234567890},
		want:  "QmPSiAdbBuuGiziorGGbjxJta9mHjhF1AhY7npLQNS9urf",
		json:  `[1234567890,1234567890]`,
	}, {
		name:  "af",
		value: chore.FloatArray{12345.6789, 12345.6789},
		want:  "QmfDekVnxZscWAtrWBqFvRjp7dfnJysrDP9Qt5a3gndz1f",
		json:  `[12345.6789,12345.6789]`,
	}, {
		name:  "au",
		value: chore.UintArray{123456789, 123456789},
		want:  "QmeicuSaCmkxKUXcDaPL5UYdkLd5AuwkCWBttwFfQSwXRR",
		json:  `[123456789,123456789]`,
	}, {
		name:  "ah",
		value: chore.CIDArray{dummyCID, dummyCID},
		want:  "QmR5EZ8hSCuxeRVpQ3HEK4sfu257j3TDk4aAjGsPYYSq3h",
		json:  `["bahw5yaisecyj3ylr734qllogfe5zo7vyjdqo7iignjdce35gh4kccr37ot5qu","bahw5yaisecyj3ylr734qllogfe5zo7vyjdqo7iignjdce35gh4kccr37ot5qu"]`,
	}, {
		name: "m>s",
		value: chore.Map{
			"foo": chore.String("foo"),
		},
		want: "QmdHFkHepSqpW2miv4YmHPebFSTUx3ksQ8Vq673MUiq7Hg",
		json: `{"foo:s":"foo"}`,
	}, {
		name: "am>s",
		value: chore.MapArray{
			chore.Map{
				"foo": chore.String("foo"),
			},
			chore.Map{
				"foo": chore.String("foo"),
			},
		},
		want: "QmV1hN4R53trE28xJxoLgUvHJFxbnjDQZ7oJwkTY4uKZCx",
		json: `[{"foo:s":"foo"},{"foo:s":"foo"}]`,
	}, {
		name: "m>d",
		value: chore.Map{
			"foo": chore.Data([]byte{0, 1, 2}),
		},
		want: "QmQw1HF2SczQQfQHgoqMAX2KsMedM6ZmEJgrRqHdM9o2qR",
		json: `{"foo:d":"AAEC"}`,
	}, {
		name: "m>b0",
		value: chore.Map{
			"foo": chore.Bool(false),
		},
		want: "QmafqTRqHM61GYdn5v7GwtxtcjPHck2t1Kkr4emW6gwNXS",
		json: `{"foo:b":false}`,
	}, {
		name: "m>b1",
		value: chore.Map{
			"foo": chore.Bool(true),
		},
		want: "QmP5uMch8FzYYEKL8KoWP3dMwSe1fhxjAjk3U9qaVXXqHZ",
		json: `{"foo:b":true}`,
	}, {
		name: "m>i",
		value: chore.Map{
			"foo": chore.Int(1234567890),
		},
		want: "QmZBnSMnaB78XxSeNjMu6w8kuiG4NR9gUBSztetDd5MHfA",
		json: `{"foo:i":1234567890}`,
	}, {
		name: "m>f",
		value: chore.Map{
			"foo": chore.Float(12345.6789),
		},
		want: "QmQBjgRdKMaDSPvR2kAsU1CQRuXqgANPFBh5hpa2qgKJYh",
		json: `{"foo:f":12345.6789}`,
	}, {
		name:  "m",
		value: dummy,
		want:  "QmaE66T7meLdCVSqpsiau24GKAttsz3sNeVQutsWprU2m7",
		json:  `{"@type:s":"dummy","foo:s":"bar"}`,
	}, {
		name: "m>o",
		value: chore.Map{
			"foo": dummy,
		},
		want: "QmWxsrko9Lv7moWKfiFwPooKm9o6pDyxv3iv5xnLETb4kj",
		json: `{"foo:m":{"@type:s":"dummy","foo:s":"bar"}}`,
	}, {
		name: "m>h",
		value: chore.Map{
			"foo": dummyCID,
		},
		want: "QmWxsrko9Lv7moWKfiFwPooKm9o6pDyxv3iv5xnLETb4kj",
		json: `{"foo:r":"bahw5yaisecyj3ylr734qllogfe5zo7vyjdqo7iignjdce35gh4kccr37ot5qu"}`,
	}, {
		name: "m>_sig",
		value: chore.Map{
			"foo":        dummy,
			"_signature": chore.String("should not matter"),
		},
		want: "QmWxsrko9Lv7moWKfiFwPooKm9o6pDyxv3iv5xnLETb4kj",
		json: `{"_signature:s":"should not matter","foo:m":{"@type:s":"dummy","foo:s":"bar"}}`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test marshaling
			b, err := json.Marshal(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.json, string(b))
			// test unmarshaling of maps
			// TODO add the remaining types by reflecting tt.value
			if _, ok := tt.value.(chore.Map); ok {
				m := chore.Map{}
				err = json.Unmarshal([]byte(tt.json), &m)
				require.NoError(t, err)
				assert.Equal(t, tt.value, m)
			}
			// test multihashes
			got, err := cid.FromValue(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.B58String())
		})
	}
}

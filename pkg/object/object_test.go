package object

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/object/cid"
	"nimona.io/pkg/object/value"
)

func TestObject(t *testing.T) {
	o := value.Map{
		"boolArray": value.BoolArray{
			value.Bool(false),
			value.Bool(true),
		},
		"dataArray": value.DataArray{
			value.Data("v0"),
			value.Data("v1"),
		},
		"floatArray": value.FloatArray{
			value.Float(0.10),
			value.Float(1.12),
		},
		"intArray": value.IntArray{
			value.Int(0),
			value.Int(1),
		},
		"mapArray": value.MapArray{
			value.Map{"foo0": value.String("bar0")},
			value.Map{
				"boolArray": value.BoolArray{
					value.Bool(false),
					value.Bool(true),
				},
				"dataArray": value.DataArray{
					value.Data("v0"),
					value.Data("v1"),
				},
				"floatArray": value.FloatArray{
					value.Float(0.10),
					value.Float(1.12),
				},
				"intArray": value.IntArray{
					value.Int(0),
					value.Int(1),
				},
				"mapArray": value.MapArray{
					value.Map{"foo0": value.String("bar0")},
					value.Map{"foo1": value.String("bar1")},
				},
				"stringArray": value.StringArray{
					value.String("v0"),
					value.String("v1"),
				},
				"uintArray": value.UintArray{
					value.Uint(0),
					value.Uint(1),
				},
			},
		},
		"stringArray": value.StringArray{
			value.String("v0"),
			value.String("v1"),
		},
		"uintArray": value.UintArray{
			value.Uint(0),
			value.Uint(1),
		},
		"bool":  value.Bool(true),
		"data":  value.Data("foo"),
		"float": value.Float(1.1),
		"int":   value.Int(2),
		"map": value.Map{
			"boolArray": value.BoolArray{
				value.Bool(false),
				value.Bool(true),
			},
			"dataArray": value.DataArray{
				value.Data("v0"),
				value.Data("v1"),
			},
			"floatArray": value.FloatArray{
				value.Float(0.10),
				value.Float(1.12),
			},
			"value.IntArray": value.IntArray{
				value.Int(0),
				value.Int(1),
			},
			"mapArray": value.MapArray{
				value.Map{"foo0": value.String("bar0")},
				value.Map{"foo1": value.String("bar1")},
			},
			"stringArray": value.StringArray{
				value.String("v0"),
				value.String("v1"),
			},
			"value.UintArray": value.UintArray{
				value.Uint(0),
				value.Uint(1),
			},
			"bool":  value.Bool(true),
			"data":  value.Data("foo"),
			"float": value.Float(1.1),
			"int":   value.Int(2),
			"map": value.Map{
				"int": value.Int(42),
			},
			"string": value.String("foo"),
			"uint":   value.Uint(3),
		},
		"string": value.String("foo"),
		"uint":   value.Uint(3),
	}
	b, err := json.MarshalIndent(o, "", "  ")
	require.NoError(t, err)

	g := value.Map{}
	err = json.Unmarshal(b, &g)
	require.NoError(t, err)

	require.Equal(t, o, g)
}

// nolint: lll
func TestValues(t *testing.T) {
	dummyObj := &Object{
		Type: "dummy",
		Data: value.Map{
			"foo": value.String("bar"),
		},
	}
	dummy, err := dummyObj.MarshalMap()
	require.NoError(t, err)
	dummyCID, err := cid.New(dummy)
	require.NoError(t, err)
	tests := []struct {
		name    string
		value   value.Value
		json    string
		want    string
		wantErr bool
	}{{
		name:  "s",
		value: value.String("foo"),
		want:  "Qmczn3TuD5jpokERZczuuPXbgJikFgJkwgr6J9p3dnTSkj",
		json:  `"foo"`,
	}, {
		name:  "d",
		value: value.Data([]byte{0, 1, 2}),
		want:  "QmXZgMaXSSJsRg9zstbj6nURqGVppGMBKxGSUrNU8u7D2r",
		json:  `"AAEC"`,
	}, {
		name:  "b0",
		value: value.Bool(false),
		want:  "QmQP6XKQetJQ1ZQNm5SoC7skxj28J4QAGhSnrHkEA2gNWU",
		json:  `false`,
	}, {
		name:  "b1",
		value: value.Bool(true),
		want:  "QmPR2X9b88aGKUTEh7UuFdcvG81RQ72jXUsMHob1qRZFwa",
		json:  `true`,
	}, {
		name:  "i",
		value: value.Int(1234567890),
		want:  "QmcaJDHHUUsiwcYk554TCLMQxvDYvgdH9f1n4MAM9JuqC9",
		json:  `1234567890`,
	}, {
		name:  "f",
		value: value.Float(12345.6789),
		want:  "QmdMKevPGKgWFGJCfmWxmMWQpztKHRo9kmF6dyx6MWuSmZ",
		json:  `12345.6789`,
	}, {
		name:  "u",
		value: value.Uint(123456789),
		want:  "QmanqLLrcWahzVH4VUKsfq44VegJSS6G1gYk5cLUSCYK9x",
		json:  `123456789`,
	}, {
		name:  "as",
		value: value.StringArray{"foo", "foo"},
		want:  "QmW2K4k56Fxvyk7v6YQ4iEeKtNeCd2m2Nu1xwPQtSR6fho",
		json:  `["foo","foo"]`,
	}, {
		name:  "ad",
		value: value.DataArray{[]byte{0, 1, 2}, []byte{0, 1, 2}},
		want:  "QmeBcPzgGU1YGdSY5SCr4CVMnA2bopJ4tUN5McaPZVkyi1",
		json:  `["AAEC","AAEC"]`,
	}, {
		name:  "ab0",
		value: value.BoolArray{false, false},
		want:  "QmcC2yFC2FvrZVpgahacSmBSx34sPcY2Tnh3VSfGgsZjhH",
		json:  `[false,false]`,
	}, {
		name:  "ab1",
		value: value.BoolArray{true, true},
		want:  "QmXqhrACfxuhKfKoCQegJwrMixxmut84kJcJpiZ8tiSXYi",
		json:  `[true,true]`,
	}, {
		name:  "ai",
		value: value.IntArray{1234567890, 1234567890},
		want:  "QmPSiAdbBuuGiziorGGbjxJta9mHjhF1AhY7npLQNS9urf",
		json:  `[1234567890,1234567890]`,
	}, {
		name:  "af",
		value: value.FloatArray{12345.6789, 12345.6789},
		want:  "QmfDekVnxZscWAtrWBqFvRjp7dfnJysrDP9Qt5a3gndz1f",
		json:  `[12345.6789,12345.6789]`,
	}, {
		name:  "au",
		value: value.UintArray{123456789, 123456789},
		want:  "QmeicuSaCmkxKUXcDaPL5UYdkLd5AuwkCWBttwFfQSwXRR",
		json:  `[123456789,123456789]`,
	}, {
		name:  "ah",
		value: value.CIDArray{dummyCID, dummyCID},
		want:  "QmR5EZ8hSCuxeRVpQ3HEK4sfu257j3TDk4aAjGsPYYSq3h",
		json:  `["bahw5yaisecyj3ylr734qllogfe5zo7vyjdqo7iignjdce35gh4kccr37ot5qu","bahw5yaisecyj3ylr734qllogfe5zo7vyjdqo7iignjdce35gh4kccr37ot5qu"]`,
	}, {
		name: "m>s",
		value: value.Map{
			"foo": value.String("foo"),
		},
		want: "QmdHFkHepSqpW2miv4YmHPebFSTUx3ksQ8Vq673MUiq7Hg",
		json: `{"foo:s":"foo"}`,
	}, {
		name: "am>s",
		value: value.MapArray{
			value.Map{
				"foo": value.String("foo"),
			},
			value.Map{
				"foo": value.String("foo"),
			},
		},
		want: "QmV1hN4R53trE28xJxoLgUvHJFxbnjDQZ7oJwkTY4uKZCx",
		json: `[{"foo:s":"foo"},{"foo:s":"foo"}]`,
	}, {
		name: "m>d",
		value: value.Map{
			"foo": value.Data([]byte{0, 1, 2}),
		},
		want: "QmQw1HF2SczQQfQHgoqMAX2KsMedM6ZmEJgrRqHdM9o2qR",
		json: `{"foo:d":"AAEC"}`,
	}, {
		name: "m>b0",
		value: value.Map{
			"foo": value.Bool(false),
		},
		want: "QmafqTRqHM61GYdn5v7GwtxtcjPHck2t1Kkr4emW6gwNXS",
		json: `{"foo:b":false}`,
	}, {
		name: "m>b1",
		value: value.Map{
			"foo": value.Bool(true),
		},
		want: "QmP5uMch8FzYYEKL8KoWP3dMwSe1fhxjAjk3U9qaVXXqHZ",
		json: `{"foo:b":true}`,
	}, {
		name: "m>i",
		value: value.Map{
			"foo": value.Int(1234567890),
		},
		want: "QmZBnSMnaB78XxSeNjMu6w8kuiG4NR9gUBSztetDd5MHfA",
		json: `{"foo:i":1234567890}`,
	}, {
		name: "m>f",
		value: value.Map{
			"foo": value.Float(12345.6789),
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
		value: value.Map{
			"foo": dummy,
		},
		want: "QmWxsrko9Lv7moWKfiFwPooKm9o6pDyxv3iv5xnLETb4kj",
		json: `{"foo:m":{"@type:s":"dummy","foo:s":"bar"}}`,
	}, {
		name: "m>h",
		value: value.Map{
			"foo": dummyCID,
		},
		want: "QmWxsrko9Lv7moWKfiFwPooKm9o6pDyxv3iv5xnLETb4kj",
		json: `{"foo:r":"bahw5yaisecyj3ylr734qllogfe5zo7vyjdqo7iignjdce35gh4kccr37ot5qu"}`,
	}, {
		name: "m>_sig",
		value: value.Map{
			"foo":        dummy,
			"_signature": value.String("should not matter"),
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
			if _, ok := tt.value.(value.Map); ok {
				m := value.Map{}
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

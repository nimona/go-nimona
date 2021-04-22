package object

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	o := Map{
		"boolArray": BoolArray{
			Bool(false),
			Bool(true),
		},
		"dataArray": DataArray{
			Data("v0"),
			Data("v1"),
		},
		"floatArray": FloatArray{
			Float(0.10),
			Float(1.12),
		},
		"intArray": IntArray{
			Int(0),
			Int(1),
		},
		"mapArray": MapArray{
			Map{"foo0": String("bar0")},
			Map{
				"boolArray": BoolArray{
					Bool(false),
					Bool(true),
				},
				"dataArray": DataArray{
					Data("v0"),
					Data("v1"),
				},
				"floatArray": FloatArray{
					Float(0.10),
					Float(1.12),
				},
				"intArray": IntArray{
					Int(0),
					Int(1),
				},
				"mapArray": MapArray{
					Map{"foo0": String("bar0")},
					Map{"foo1": String("bar1")},
				},
				"stringArray": StringArray{
					String("v0"),
					String("v1"),
				},
				"uintArray": UintArray{
					Uint(0),
					Uint(1),
				},
			},
		},
		"stringArray": StringArray{
			String("v0"),
			String("v1"),
		},
		"uintArray": UintArray{
			Uint(0),
			Uint(1),
		},
		"bool":  Bool(true),
		"data":  Data("foo"),
		"float": Float(1.1),
		"int":   Int(2),
		"map": Map{
			"boolArray": BoolArray{
				Bool(false),
				Bool(true),
			},
			"dataArray": DataArray{
				Data("v0"),
				Data("v1"),
			},
			"floatArray": FloatArray{
				Float(0.10),
				Float(1.12),
			},
			"IntArray": IntArray{
				Int(0),
				Int(1),
			},
			"mapArray": MapArray{
				Map{"foo0": String("bar0")},
				Map{"foo1": String("bar1")},
			},
			"stringArray": StringArray{
				String("v0"),
				String("v1"),
			},
			"UintArray": UintArray{
				Uint(0),
				Uint(1),
			},
			"bool":  Bool(true),
			"data":  Data("foo"),
			"float": Float(1.1),
			"int":   Int(2),
			"map": Map{
				"int": Int(42),
			},
			"string": String("foo"),
			"uint":   Uint(3),
		},
		"string": String("foo"),
		"uint":   Uint(3),
	}
	b, err := json.MarshalIndent(o, "", "  ")
	require.NoError(t, err)

	g := Map{}
	err = json.Unmarshal(b, &g)
	require.NoError(t, err)

	require.Equal(t, o, g)
}

// nolint: lll
func TestValues(t *testing.T) {
	dummy := &Object{
		Type: "dummy",
		Data: Map{
			"foo": String("bar"),
		},
	}
	tests := []struct {
		name    string
		value   Value
		json    string
		want    string
		wantErr bool
	}{{
		name:  "s",
		value: String("foo"),
		want:  "Qmczn3TuD5jpokERZczuuPXbgJikFgJkwgr6J9p3dnTSkj",
		json:  `"foo"`,
	}, {
		name:  "d",
		value: Data([]byte{0, 1, 2}),
		want:  "QmXZgMaXSSJsRg9zstbj6nURqGVppGMBKxGSUrNU8u7D2r",
		json:  `"AAEC"`,
	}, {
		name:  "b0",
		value: Bool(false),
		want:  "QmQP6XKQetJQ1ZQNm5SoC7skxj28J4QAGhSnrHkEA2gNWU",
		json:  `false`,
	}, {
		name:  "b1",
		value: Bool(true),
		want:  "QmPR2X9b88aGKUTEh7UuFdcvG81RQ72jXUsMHob1qRZFwa",
		json:  `true`,
	}, {
		name:  "i",
		value: Int(1234567890),
		want:  "QmcaJDHHUUsiwcYk554TCLMQxvDYvgdH9f1n4MAM9JuqC9",
		json:  `1234567890`,
	}, {
		name:  "f",
		value: Float(12345.6789),
		want:  "QmdMKevPGKgWFGJCfmWxmMWQpztKHRo9kmF6dyx6MWuSmZ",
		json:  `12345.6789`,
	}, {
		name:  "u",
		value: Uint(123456789),
		want:  "QmanqLLrcWahzVH4VUKsfq44VegJSS6G1gYk5cLUSCYK9x",
		json:  `123456789`,
	}, {
		name:  "as",
		value: StringArray{"foo", "foo"},
		want:  "QmW2K4k56Fxvyk7v6YQ4iEeKtNeCd2m2Nu1xwPQtSR6fho",
		json:  `["foo","foo"]`,
	}, {
		name:  "ad",
		value: DataArray{[]byte{0, 1, 2}, []byte{0, 1, 2}},
		want:  "QmeBcPzgGU1YGdSY5SCr4CVMnA2bopJ4tUN5McaPZVkyi1",
		json:  `["AAEC","AAEC"]`,
	}, {
		name:  "ab0",
		value: BoolArray{false, false},
		want:  "QmcC2yFC2FvrZVpgahacSmBSx34sPcY2Tnh3VSfGgsZjhH",
		json:  `[false,false]`,
	}, {
		name:  "ab1",
		value: BoolArray{true, true},
		want:  "QmXqhrACfxuhKfKoCQegJwrMixxmut84kJcJpiZ8tiSXYi",
		json:  `[true,true]`,
	}, {
		name:  "ai",
		value: IntArray{1234567890, 1234567890},
		want:  "QmPSiAdbBuuGiziorGGbjxJta9mHjhF1AhY7npLQNS9urf",
		json:  `[1234567890,1234567890]`,
	}, {
		name:  "af",
		value: FloatArray{12345.6789, 12345.6789},
		want:  "QmfDekVnxZscWAtrWBqFvRjp7dfnJysrDP9Qt5a3gndz1f",
		json:  `[12345.6789,12345.6789]`,
	}, {
		name:  "au",
		value: UintArray{123456789, 123456789},
		want:  "QmeicuSaCmkxKUXcDaPL5UYdkLd5AuwkCWBttwFfQSwXRR",
		json:  `[123456789,123456789]`,
	}, {
		name:  "ah",
		value: CIDArray{dummy.CID(), dummy.CID()},
		want:  "QmVHnb5wNonYoaPyqpcJkJcu38QJpW3dNk2MiXBCL3zBtT",
		json:  `["bahw5yaisecl5762aft2rdcnsockiea3nxjpdnvbbys5dnbho77svapvwsagwi","bahw5yaisecl5762aft2rdcnsockiea3nxjpdnvbbys5dnbho77svapvwsagwi"]`,
	}, {
		name: "m>s",
		value: Map{
			"foo": String("foo"),
		},
		want: "QmY4Muu5Cc5ACMA7aHNAUsGkAZ3jXXG43sT27UpVpMSUeD",
		json: `{"foo:s":"foo"}`,
	}, {
		name: "am>s",
		value: MapArray{
			Map{
				"foo": String("foo"),
			},
			Map{
				"foo": String("foo"),
			},
		},
		want: "QmXH6Jo5J9m8oVNA2aujuK3aRUZoEo5HVdftoGQvFTuD4V",
		json: `[{"foo:s":"foo"},{"foo:s":"foo"}]`,
	}, {
		name: "m>d",
		value: Map{
			"foo": Data([]byte{0, 1, 2}),
		},
		want: "QmdcWPVivKuAnLbhDhX2dg5nzEhvaZ11fCm4L7XtXWgbjS",
		json: `{"foo:d":"AAEC"}`,
	}, {
		name: "m>b0",
		value: Map{
			"foo": Bool(false),
		},
		want: "QmU91cYR7ELzxnxs22uqoHT9cA7pZx1Xvd1N6RyKpQ1wid",
		json: `{"foo:b":false}`,
	}, {
		name: "m>b1",
		value: Map{
			"foo": Bool(true),
		},
		want: "QmP1Tx5dmezZhjwqaFGJv5e6WsYnoQu4GicZC28SAai8t1",
		json: `{"foo:b":true}`,
	}, {
		name: "m>i",
		value: Map{
			"foo": Int(1234567890),
		},
		want: "QmewfnPYHdz4ztxCaR8sXtjEjYqhdLUH89VzggieoRcYVV",
		json: `{"foo:i":1234567890}`,
	}, {
		name: "m>f",
		value: Map{
			"foo": Float(12345.6789),
		},
		want: "QmcuTf34Mj8To8sVJRfNeGEMD8aoRK5rNewcSHjEScSuXc",
		json: `{"foo:f":12345.6789}`,
	}, {
		name:  "o",
		value: dummy,
		want:  "QmYZWMxneHYqxSoZHr45cXPLGPzRrLLdgDsCAqc7XnFWCK",
		json:  `{"@type:s":"dummy","foo:s":"bar"}`,
	}, {
		name: "ao",
		value: ObjectArray{
			dummy,
			dummy,
		},
		want: "QmY9QbAQ2kJ67tms5t63QWPjXQ5pB5Zb7nsUa6UcTtCsxX",
		json: `[{"@type:s":"dummy","foo:s":"bar"},{"@type:s":"dummy","foo:s":"bar"}]`,
	}, {
		name: "m>o",
		value: Map{
			"foo": dummy,
		},
		want: "QmR7sCawpQ2Eu6BkSnjRKKLzuRmR6PXhxBrHmnzHfKSbVk",
		json: `{"foo:o":{"@type:s":"dummy","foo:s":"bar"}}`,
	}, {
		name: "m>h",
		value: Map{
			"foo": dummy.CID(),
		},
		want: "QmR7sCawpQ2Eu6BkSnjRKKLzuRmR6PXhxBrHmnzHfKSbVk",
		json: `{"foo:r":"bahw5yaisecl5762aft2rdcnsockiea3nxjpdnvbbys5dnbho77svapvwsagwi"}`,
	}, {
		name: "m>_sig",
		value: Map{
			"foo":        dummy,
			"_signature": String("should not matter"),
		},
		want: "QmR7sCawpQ2Eu6BkSnjRKKLzuRmR6PXhxBrHmnzHfKSbVk",
		json: `{"_signature:s":"should not matter","foo:o":{"@type:s":"dummy","foo:s":"bar"}}`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test marshaling
			b, err := json.Marshal(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.json, string(b))
			// test unmarshaling of maps
			// TODO add the remaining types by reflecting tt.value
			if _, ok := tt.value.(Map); ok {
				m := Map{}
				err = json.Unmarshal([]byte(tt.json), &m)
				require.NoError(t, err)
				assert.Equal(t, tt.value, m)
			}
			// test multihashes
			got, err := fromValue(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.B58String())
		})
	}
}

package tilde

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkDigest(b *testing.B) {
	o := Map{
		"@type":    String("blob"),
		"filename": String("foo"),
		"dummy": Map{
			"@type": String("dummy"),
			"@metadata": Map{
				"owner": String("foo"),
			},
			"foo": String("bar"),
			"data": Data(
				"1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14" +
					"\n15\n16\n17\n18\n19\n20",
			),
		},
	}
	for n := 0; n < b.N; n++ {
		o.Hash()
	}
}

func TestMapRedaction(t *testing.T) {
	inner := Map{
		"@type": String("foo"),
		"foo":   String("bar"),
	}
	parentWithInner := Map{
		"@type": String("foo"),
		"foo":   inner,
	}
	parentWithInnerRedacted := Map{
		"@type": String("foo"),
		"foo":   inner.Hash(),
	}
	tests := []struct {
		name       string
		json       string
		wantValue  Value
		wantDigest Digest
		wantErr    bool
	}{{
		name:       "1",
		json:       `{"@type:s":"foo","foo:s":"bar"}`,
		wantValue:  inner,
		wantDigest: inner.Hash(),
	}, {
		name:       "2",
		json:       `{"@type:s":"foo","foo:m":{"@type:s":"foo","foo:s":"bar"}}`,
		wantValue:  parentWithInner,
		wantDigest: parentWithInner.Hash(),
	}, {
		name:       "3",
		json:       `{"@type:s":"foo","foo:r":"` + string(inner.Hash()) + `"}`,
		wantValue:  parentWithInnerRedacted,
		wantDigest: parentWithInner.Hash(),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Map{}
			fmt.Println(tt.json)
			assert.NoError(t, json.Unmarshal([]byte(tt.json), &m))
			assert.Equal(t, tt.wantValue, m)
			got := m.Hash()
			assert.Equal(t, tt.wantDigest, got)
		})
	}
}

// nolint: lll
func TestValues(t *testing.T) {
	dummy := Map{
		"@type": String("dummy"),
		"foo":   String("bar"),
	}
	dummyDigest := dummy.Hash()
	tests := []struct {
		name    string
		value   Value
		json    string
		want    Digest
		wantErr bool
	}{{
		name:  "s",
		value: String("foo"),
		want:  "3yMApqCuCjXDWPrbjfR5mjCPTHqFG8Pux1TxQrEM35jj",
		json:  `"foo"`,
	}, {
		name:  "d",
		value: Data([]byte{0, 1, 2}),
		want:  "CjNSmWXTWhC3EhRVtqLhRmWMTkRbU96wUACqxMtV1uGf",
		json:  `"AAEC"`,
	}, {
		name:  "b0",
		value: Bool(false),
		want:  "8RBsoeyoRwajj86MZfZE6gMDJQVYGYcdSfx1zxqxNHbr",
		json:  `false`,
	}, {
		name:  "b1",
		value: Bool(true),
		want:  "67WKXSxm4oc149PvQjdXLacKFZpK5DyYdqBwpiVydJbb",
		json:  `true`,
	}, {
		name:  "i",
		value: Int(1234567890),
		want:  "ERcQehiQ6dbLV96CpJKu8nAKYz5LGPSy9ndt1JAtbzkd",
		json:  `1234567890`,
	}, {
		name:  "f",
		value: Float(12345.6789),
		want:  "DS8LSZA5GygBnCF3evGv18UQhWqNNwcwXAEQQ5ycuPn",
		json:  `12345.6789`,
	}, {
		name:  "u",
		value: Uint(123456789),
		want:  "2US3nkwKL3JYXoWu3jq6TZGh1Cm1S6o8vkLAEoK4cELU",
		json:  `123456789`,
	}, {
		name:  "r",
		value: Digest("bQbp"),
		want:  "bQbp",
		json:  `"bQbp"`,
	}, {
		name:  "as",
		value: StringArray{"foo", "foo"},
		want:  "86mB69kbuwhxVrT2qZ9z4GQEB7YVFRKLniXGp3uYNaqH",
		json:  `["foo","foo"]`,
	}, {
		name:  "ad",
		value: DataArray{[]byte{0, 1, 2}, []byte{0, 1, 2}},
		want:  "HjjiMUwg5fqqczSUJrWETWLxWo5Z5H5W4KscbZ5mzDmi",
		json:  `["AAEC","AAEC"]`,
	}, {
		name:  "ab0",
		value: BoolArray{false, false},
		want:  "36VXqdzF8RgzyJggTBtED6oz2N82Lc81eLaTMdG5RM7h",
		json:  `[false,false]`,
	}, {
		name:  "ab1",
		value: BoolArray{true, true},
		want:  "Av2FzunhMgiFEBLXUvbEaWgFJsGwAGk1BhMNeSYeH4G3",
		json:  `[true,true]`,
	}, {
		name:  "ai",
		value: IntArray{1234567890, 1234567890},
		want:  "493FLheumd5YEjbj7VeZ1euLqw2sXsEXRPd4mwDGMkjy",
		json:  `[1234567890,1234567890]`,
	}, {
		name:  "af",
		value: FloatArray{12345.6789, 12345.6789},
		want:  "4nYrYjHQv5P66bi7TCsjtYNsk3ti7en6cbGkBbYKxRce",
		json:  `[12345.6789,12345.6789]`,
	}, {
		name:  "au",
		value: UintArray{123456789, 123456789},
		want:  "8DR5Y5aBcJL6TxHffFeDmTFKwMZgGqotSBaPdpUbgU3c",
		json:  `[123456789,123456789]`,
	}, {
		name:  "ah",
		value: DigestArray{dummyDigest, dummyDigest},
		want:  "CJPJ9tJF5HUGXxCeH6QPiANMSrxSZua4bKABuwnYkL63",
		json:  `["8ik1XwTQrmt7punAKVz3AZ9HTCp9FFqocUVKre1GrVKU","8ik1XwTQrmt7punAKVz3AZ9HTCp9FFqocUVKre1GrVKU"]`,
	}, {
		name: "m>s",
		value: Map{
			"foo": String("foo"),
		},
		want: "5RyEnsUHRUwdjbjqmw9VGD1sP3SHGrrWTfiZyuAC8URw",
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
		want: "CPzthKrxSzqDuJ4HixwH8Rb8jnwD8PgNbSuC2r7QvJTK",
		json: `[{"foo:s":"foo"},{"foo:s":"foo"}]`,
	}, {
		name: "m>d",
		value: Map{
			"foo": Data([]byte{0, 1, 2}),
		},
		want: "D2jWDeSFSTfb9MUvvoff4yRZtf7ExV4HGt6APfPc2APd",
		json: `{"foo:d":"AAEC"}`,
	}, {
		name: "m>b0",
		value: Map{
			"foo": Bool(false),
		},
		want: "EqHEui7jzdEmt4moGzHvi9bE7T9nfPQmcFADJR2mcWEh",
		json: `{"foo:b":false}`,
	}, {
		name: "m>b1",
		value: Map{
			"foo": Bool(true),
		},
		want: "9uQoB7ibtmKRBu5mDEwim5KNA3A1cotm3bHsFns2UWS6",
		json: `{"foo:b":true}`,
	}, {
		name: "m>i",
		value: Map{
			"foo": Int(1234567890),
		},
		want: "5nku7DvovAg2PsNwi88pFXxXS5jPgxWeVSn77vRXwTZ9",
		json: `{"foo:i":1234567890}`,
	}, {
		name: "m>f",
		value: Map{
			"foo": Float(12345.6789),
		},
		want: "5jgpbKQPtnaMJ9wZf1hH8brWQszCEk3waLitdM4Zvj6p",
		json: `{"foo:f":12345.6789}`,
	}, {
		name:  "m",
		value: dummy,
		want:  "8ik1XwTQrmt7punAKVz3AZ9HTCp9FFqocUVKre1GrVKU",
		json:  `{"@type:s":"dummy","foo:s":"bar"}`,
	}, {
		name: "m>m",
		value: Map{
			"foo": dummy,
		},
		want: "7wTd5d7KMQrSAZZLKBSgRwstDRrgwpUz2fwRAAFyUozr",
		json: `{"foo:m":{"@type:s":"dummy","foo:s":"bar"}}`,
	}, {
		name: "m>h",
		value: Map{
			"foo": dummyDigest,
		},
		want: "7wTd5d7KMQrSAZZLKBSgRwstDRrgwpUz2fwRAAFyUozr",
		json: `{"foo:r":"8ik1XwTQrmt7punAKVz3AZ9HTCp9FFqocUVKre1GrVKU"}`,
	}, {
		name: "m>_sig",
		value: Map{
			"foo":        dummy,
			"_signature": String("should not matter"),
		},
		want: "7wTd5d7KMQrSAZZLKBSgRwstDRrgwpUz2fwRAAFyUozr",
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
			if _, ok := tt.value.(Map); ok {
				m := Map{}
				err = json.Unmarshal([]byte(tt.json), &m)
				require.NoError(t, err)
				assert.Equal(t, tt.value, m)
			}
			got := tt.value.Hash()
			assert.Equal(t, tt.want, got)
		})
	}
}

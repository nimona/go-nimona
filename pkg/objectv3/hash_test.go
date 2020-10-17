package objectv3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHash(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		json    string
		want    Hash
		wantErr bool
	}{{
		key:  ":s",
		json: `"foo"`,
		want: "oh1.Ff8LMmDZxqL7dn9SxiBusRxKDGQ5f7NzmnHJ7tEUFkgj",
	}, {
		key:  ":d",
		json: `"Zm9v"`,
		want: "oh1.EuzQfFdCYuqmzZ5Htk56VzmYPYxdfbqPchqeps5d4q85",
	}, {
		key:  ":b",
		json: `false`,
		want: "oh1.33SpDGfNXQXvojLuQbUeDbEjWeSTkWi1NUqS44kiUgSU",
	}, {
		key:  ":b",
		json: `true`,
		want: "oh1.25Np3T8coGqqrbGwShY9xkY8VwYWP5xno3GwtrS7MZsa",
	}, {
		key:  ":i",
		json: `1234567890`,
		want: "oh1.FEeWB9Uy6jTyx6eu2FUrhFEvi555ydaxwU3VUBjzi989",
	}, {
		key:  ":f",
		json: `12345.6789`,
		want: "oh1.6YbTAHuuTfan9z8j6UVG9MYH8Y9oy9u6ZpiroFXoqXkT",
	}, {
		key:  ":m",
		json: `{"something:s":"foo"}`,
		want: "oh1.3GFaM2nhTSuEUGh29tFdktAH1u79mCFKTD3NAzwMTUVf",
	}, {
		key:  ":m",
		json: `{"something:d":"Zm9v"}`,
		want: "oh1.FHMpNW8UHavirmi6Ag9sFSVVJpNT2rgtxSTnxjmZufQ",
	}, {
		key:  ":m",
		json: `{"something:b":false}`,
		want: "oh1.FUdMm76daGuwJmMrDn2Uzb8U7Y8xSipofSSJRbKcXG4b",
	}, {
		key:  ":m",
		json: `{"something:b":true}`,
		want: "oh1.ECjyVPPabfQpqD4zdviWbVmh44VtDQrfdeiT14P7d2Kh",
	}, {
		key:  ":m",
		json: `{"something:i":1234567890}`,
		want: "oh1.DveWCsBTBBNwGmmTeZbpCedrME1E2XHe69J9PGZceDJm",
	}, {
		key:  ":m",
		json: `{"something:f":12345.6789}`,
		want: "oh1.4XiNVLLVyD3yAyZzpk7kCC6SW8t5hYaMWNK7JyGmsigJ",
	}, {
		key:  ":as",
		json: `["foo","bar"]`,
		want: "oh1.3LGHmZJpypMyxdwWtdva89cydgqfJU5W12cTrEB6erHb",
	}, {
		key:  ":m",
		json: `{"something:as":["foo","bar"]}`,
		want: "oh1.5ttZgwrbiVeERiQ17YMXbTsHP3NvLRRqsfYeNEnfxvgq",
	}, {
		key:  ":m",
		json: `{"something:ai":[123,456]}`,
		want: "oh1.GvhANQSTivbTre6UmkBEhVMFo3aGyhXscKmHCg7Nm4tT",
	}, {
		key:  ":m",
		json: `{"foo:s":"bar"}`,
		want: "oh1.CgfoHRELcu1DwPjtGcXuVr1oFbAVxF8mRTWkTyJsE9gk",
	}, {
		key:  ":m",
		json: `{"data:r":"oh1.CgfoHRELcu1DwPjtGcXuVr1oFbAVxF8mRTWkTyJsE9gk"}`,
		want: "oh1.EAKxMZySQigLYF9hZ3D4YjqrhWQ6q24NhvvbmUAsQSCt",
	}, {
		key:  ":m",
		json: `{"data:m":{"foo:s":"bar","nested:o":{"_sig:s":"should not matter","foo:s":"bar"}}}`,
		want: "oh1.CA3EJnaqMXGVZuMzb5DS2vTUBcjjwwsikyAFtQA1uLQm",
	}, {
		key:  ":m",
		json: `{"data:m":{"foo:s":"bar","nested:o":{"_signature:m":{"foo:s":"bar"},"foo:s":"bar"}}}`,
		want: "oh1.CA3EJnaqMXGVZuMzb5DS2vTUBcjjwwsikyAFtQA1uLQm",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m interface{}
			require.NoError(t, json.Unmarshal([]byte(tt.json), &m))
			got, err := hashValueAs(tt.key, m, hintsFromKey(tt.key)...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

package objectv3

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	Chunk struct {
		Metadata Metadata `nimona:"metadata:m"`
		Index    int      `nimona:"index:i,omitempty"`
		Data     string   `nimona:"data:s,omitempty"`
	}
	Blob struct {
		Metadata Metadata `nimona:"metadata:m"`
		Dummy    *Dummy   `nimona:"dummy:m,omitempty"`
		Filename string   `nimona:"filename:s,omitempty"`
		Chunks   []*Chunk `nimona:"chunks:am,omitempty"`
	}
	Dummy struct {
		Metadata Metadata `nimona:"metadata:m"`
		Foo      string   `nimona:"foo:s,omitempty"`
	}
	// BlobUnloaded struct {
	// 	Metadata    object.Metadata `nimona:",metadata"`
	// 	Filename    string          `nimona:"filename,string"`
	// 	ChunkHashes []object.Hash   `nimona:"chunks,object"`
	// }
	// BlobMaybeUnloaded struct {
	// 	Metadata    object.Metadata           `nimona:",metadata"`
	// 	Filename    string                    `nimona:"filename,string"`
	// 	ChunkHashes []object.ResolvableObject `nimona:"chunks,object"`
	// }
)

func (v Dummy) Type() string {
	return "dummy"
}

func (v Chunk) Type() string {
	return "chunk"
}

func (v Blob) Type() string {
	return "blob"
}

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name       string
		source     interface{}
		object     *Object
		encodeOnly bool
		wantErr    bool
	}{{
		name: "object to struct, encode-decode",
		source: &Chunk{
			Metadata: Metadata{
				Owner: "foo",
			},
			Index: 1,
		},
		object: &Object{
			Type: "chunk",
			Metadata: Metadata{
				Owner: "foo",
			},
			Data: map[string]interface{}{
				"index:i": int64(1),
			},
		},
	}, {
		name: "map[string]interface{} to struct, encode",
		source: map[string]interface{}{
			"type:s": "chunk",
			"metadata:m": map[string]interface{}{
				"owner:s": "foo",
			},
			"data:m": map[string]interface{}{
				"index:i": 1,
			},
		},
		object: &Object{
			Type: "chunk",
			Metadata: Metadata{
				Owner: "foo",
			},
			Data: map[string]interface{}{
				"index:i": int64(1),
			},
		},
		encodeOnly: true,
	}, {
		name: "map[interface{}]interface{} to struct, encode",
		source: map[interface{}]interface{}{
			"type:s": "chunk",
			"metadata:m": map[interface{}]interface{}{
				"owner:s": "foo",
			},
			"data:m": map[interface{}]interface{}{
				"index:i": 1,
			},
		},
		object: &Object{
			Type: "chunk",
			Metadata: Metadata{
				Owner: "foo",
			},
			Data: map[string]interface{}{
				"index:i": int64(1),
			},
		},
		encodeOnly: true,
	}, {
		name: "json to struct, encode",
		source: func() interface{} {
			s := map[string]interface{}{
				"type:s": "chunk",
				"metadata:m": map[string]interface{}{
					"owner:s": "foo",
				},
				"data:m": map[string]interface{}{
					"index:i": 1,
				},
			}
			b, err := json.Marshal(s)
			require.NoError(t, err)
			r := map[string]interface{}{}
			err = json.Unmarshal(b, &r)
			require.NoError(t, err)
			return r
		}(),
		object: &Object{
			Type: "chunk",
			Metadata: Metadata{
				Owner: "foo",
			},
			Data: map[string]interface{}{
				"index:i": int64(1),
			},
		},
		encodeOnly: true,
	}, {
		name: "object to struct, with nested object, encode-decode",
		source: &Blob{
			Filename: "foo",
			Dummy: &Dummy{
				Metadata: Metadata{
					Owner: "foo",
				},
				Foo: "bar",
			},
		},
		object: &Object{
			Type: "blob",
			Data: map[string]interface{}{
				"filename:s": "foo",
				"dummy:m": &Object{
					Type: "dummy",
					Metadata: Metadata{
						Owner: "foo",
					},
					Data: map[string]interface{}{
						"foo:s": "bar",
					},
				},
			},
		},
	}, {
		name: "object to struct, with nested slice of objects, encode-decode",
		source: &Blob{
			Filename: "foo",
			Chunks: []*Chunk{{
				Metadata: Metadata{
					Owner: "foo",
				},
				Index: 1,
			}, {
				Metadata: Metadata{
					Owner: "foo2",
				},
				Index: 2,
			}},
		},
		object: &Object{
			Type: "blob",
			Data: map[string]interface{}{
				"filename:s": "foo",
				"chunks:am": []*Object{{
					Type: "chunk",
					Metadata: Metadata{
						Owner: "foo",
					},
					Data: map[string]interface{}{
						"index:i": int64(1),
					},
				}, {
					Type: "chunk",
					Metadata: Metadata{
						Owner: "foo2",
					},
					Data: map[string]interface{}{
						"index:i": int64(2),
					},
				}},
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			object, err := Encode(tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.object, object, "during encode")
			if !tt.encodeOnly {
				source := reflect.New(reflect.TypeOf(tt.source).Elem()).Interface().(Typed)
				err = Decode(tt.object, source)
				if (err != nil) != tt.wantErr {
					t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				assert.Equal(t, tt.source, source, "during decode")
			}
		})
	}
}

func TestDecode_ObjectWithNestedTyped(t *testing.T) {
	eb := &Blob{
		Filename: "foo",
		Dummy: &Dummy{
			Metadata: Metadata{
				Owner: "foo",
			},
			Foo: "bar",
		},
	}

	o := &Object{
		Type: "blob",
		Data: map[string]interface{}{
			"filename:s": "foo",
			"dummy:m": &Dummy{
				Metadata: Metadata{
					Owner: "foo",
				},
				Foo: "bar",
			},
		},
	}

	b := &Blob{}
	err := Decode(o, b)
	require.NoError(t, err)
	assert.Equal(t, eb, b)
}

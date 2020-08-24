package objectv3

import (
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
		Dummy    *Dummy   `nimona:"dummy:o,omitempty"`
		Filename string   `nimona:"filename:s,omitempty"`
		Chunks   []*Chunk `nimona:"chunks:ao,omitempty"`
	}
	Dummy struct {
		Metadata Metadata `nimona:"metadata:m"`
		Foo      string   `nimona:"foo:s,omitempty"`
	}
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
		name    string
		typed   Typed
		object  *Object
		wantErr bool
	}{{
		name: "single",
		typed: &Chunk{
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
				"index:i": 1,
			},
		},
	}, {
		name: "nested",
		typed: &Blob{
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
				"dummy:o": &Object{
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
		name: "slices",
		typed: &Blob{
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
				"chunks:ao": []*Object{{
					Type: "chunk",
					Metadata: Metadata{
						Owner: "foo",
					},
					Data: map[string]interface{}{
						"index:i": 1,
					},
				}, {
					Type: "chunk",
					Metadata: Metadata{
						Owner: "foo2",
					},
					Data: map[string]interface{}{
						"index:i": 2,
					},
				}},
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			object, err := Encode(tt.typed)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.object, object)
			typed := reflect.New(reflect.TypeOf(tt.typed).Elem()).Interface().(Typed)
			err = Decode(tt.object, typed)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.typed, typed)
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
			"dummy:o": &Dummy{
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

package nimona

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadata(t *testing.T) {
	o := NewTestIdentity(t).IdentityID()
	b := &DocumentBase{
		Type: "test/fixture",
		Metadata: Metadata{
			Owner: &o,
		},
	}

	t.Run("marshal unmarshal", func(t *testing.T) {
		bb, err := MarshalCBORBytes(b)
		assert.NoError(t, err)

		b1 := &DocumentBase{}
		err = UnmarshalCBORBytes(bb, b1)
		b1.DocumentBytes = nil
		assert.NoError(t, err)
		assert.EqualValues(t, b, b1)
	})
}

func TestMetadata_CheckPermissions(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		patch    StreamPatch
		want     bool
	}{{
		name: "read permission granted",
		metadata: Metadata{
			Permissions: []Permissions{{
				Paths: []string{"/path/*"},
				Operations: PermissionsAllow{
					Read: true,
				},
			}},
		},
		patch: StreamPatch{
			Operations: []StreamOperation{{
				Op:   "read",
				Path: "/path/test",
			}},
		},
		want: true,
	}, {
		name: "read permission not granted",
		metadata: Metadata{
			Permissions: []Permissions{{
				Paths: []string{"/path/*"},
				Operations: PermissionsAllow{
					Read: true,
				},
			}, {
				Paths: []string{"/path/*"},
				Operations: PermissionsAllow{
					Read: false,
				},
			}},
		},
		patch: StreamPatch{
			Operations: []StreamOperation{{
				Op:   "read",
				Path: "/path/test",
			}},
		},
		want: false,
	}, {
		name: "add permission granted",
		metadata: Metadata{
			Permissions: []Permissions{{
				Paths: []string{"/path/*"},
				Operations: PermissionsAllow{
					Add: true,
				},
			}},
		},
		patch: StreamPatch{
			Operations: []StreamOperation{{
				Op:   "add",
				Path: "/path/test",
			}},
		},
		want: true,
	}, {
		name: "add permission not granted",
		metadata: Metadata{
			Permissions: []Permissions{{
				Paths: []string{"/path/*"},
				Operations: PermissionsAllow{
					Add: false,
				},
			}},
		},
		patch: StreamPatch{
			Operations: []StreamOperation{{
				Op:   "add",
				Path: "/path/test",
			}},
		},
		want: false,
	}, {
		name: "remove permission granted",
		metadata: Metadata{
			Permissions: []Permissions{{
				Paths: []string{"/path/*"},
				Operations: PermissionsAllow{
					Remove: true,
				},
			}},
		},
		patch: StreamPatch{
			Operations: []StreamOperation{{
				Op:   "remove",
				Path: "/path/test",
			}},
		},
		want: true,
	}, {
		name: "remove permission not granted",
		metadata: Metadata{
			Permissions: []Permissions{{
				Paths: []string{"/path/*"},
				Operations: PermissionsAllow{
					Remove: false,
				},
			}},
		},
		patch: StreamPatch{
			Operations: []StreamOperation{{
				Op:   "remove",
				Path: "/path/test",
			}},
		},
		want: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CheckPermissions(&test.metadata, &test.patch)
			assert.Equal(t, test.want, got)
		})
	}
}

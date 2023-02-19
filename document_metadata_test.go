package nimona

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadata_CheckPermissions(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		patch    DocumentPatch
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
		patch: DocumentPatch{
			Operations: []DocumentPatchOperation{{
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
		patch: DocumentPatch{
			Operations: []DocumentPatchOperation{{
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
		patch: DocumentPatch{
			Operations: []DocumentPatchOperation{{
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
		patch: DocumentPatch{
			Operations: []DocumentPatchOperation{{
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
		patch: DocumentPatch{
			Operations: []DocumentPatchOperation{{
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
		patch: DocumentPatch{
			Operations: []DocumentPatchOperation{{
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

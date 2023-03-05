package nimona

import (
	"github.com/gobwas/glob"
)

type (
	Metadata struct {
		Owner       *Identity     `nimona:"owner,omitempty"`
		Permissions []Permissions `nimona:"permissions,omitempty"`
		Timestamp   uint          `nimona:"timestamp,omitempty"` // TODO: use time.Time
		Root        *DocumentID   `nimona:"root,omitempty"`
		Sequence    uint          `nimona:"sequence,omitempty"`
		Signature   *Signature    `nimona:"_signature,omitempty"`
	}
	Permissions struct {
		Paths      []string             `nimona:"paths"`
		Operations PermissionsAllow     `nimona:"operations"`
		Conditions PermissionsCondition `nimona:"conditions"`
	}
	PermissionsAllow struct {
		// generic ops
		Read bool `nimona:"read,omitempty"`
		// json patch
		Add     bool `nimona:"add,omitempty"`
		Remove  bool `nimona:"remove,omitempty"`
		Replace bool `nimona:"replace,omitempty"`
		Move    bool `nimona:"move,omitempty"`
		Copy    bool `nimona:"copy,omitempty"`
		Test    bool `nimona:"test,omitempty"`
	}
	PermissionsCondition struct {
		IsOwner bool `nimona:"isOwner,omitempty"`
	}
)

// nolint: gocyclo
func CheckPermissions(metadata *Metadata, patch *DocumentPatch) bool {
	for _, permission := range metadata.Permissions {
		patchOpAllowed := false
		for _, operation := range patch.Operations {
			pathMatches := false
			for _, path := range permission.Paths {
				g, err := glob.Compile(path, '.')
				if err != nil {
					return false
				}
				if g.Match(path) {
					pathMatches = true
					break
				}
			}
			if !pathMatches {
				continue
			}
			opMatches := false
			switch operation.Op {
			case "read":
				opMatches = permission.Operations.Read
			case "add":
				opMatches = permission.Operations.Add
			case "remove":
				opMatches = permission.Operations.Remove
			case "replace":
				opMatches = permission.Operations.Replace
			case "move":
				opMatches = permission.Operations.Move
			case "copy":
				opMatches = permission.Operations.Copy
			case "test":
				opMatches = permission.Operations.Test
			}
			if !opMatches {
				continue
			}
			if permission.Conditions.IsOwner {
				patchOpAllowed = metadata.Owner == patch.Metadata.Owner
			} else {
				patchOpAllowed = true
			}
		}
		if !patchOpAllowed {
			return false
		}
	}
	return true
}

package nimona

import (
	"github.com/gobwas/glob"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type (
	DocumentBase struct {
		Type          string   `cborgen:"$type"`
		Metadata      Metadata `cborgen:"$metadata,omitempty"`
		DocumentBytes []byte   `cborgen:"rawbytes"`
	}
	Metadata struct {
		Owner       *Identity     `cborgen:"owner,omitempty"`
		Permissions []Permissions `cborgen:"permissions,omitempty"`
		Timestamp   cbg.CborTime  `cborgen:"timestamp,omitempty"`
		Signature   Signature     `cborgen:"_signature,omitempty"`
	}
	Permissions struct {
		Paths      []string             `cborgen:"paths"`
		Operations PermissionsAllow     `cborgen:"operations"`
		Conditions PermissionsCondition `cborgen:"conditions"`
	}
	PermissionsAllow struct {
		// generic ops
		Read bool `cborgen:"read,omitempty"`
		// json patch
		Add     bool `cborgen:"add,omitempty"`
		Remove  bool `cborgen:"remove,omitempty"`
		Replace bool `cborgen:"replace,omitempty"`
		Move    bool `cborgen:"move,omitempty"`
		Copy    bool `cborgen:"copy,omitempty"`
		Test    bool `cborgen:"test,omitempty"`
	}
	PermissionsCondition struct {
		IsOwner bool `cborgen:"isOwner,omitempty"`
	}
)

// nolint: gocyclo
func CheckPermissions(metadata *Metadata, patch *DocumentPatch) bool {
	for _, permission := range metadata.Permissions {
		patchOpAllowed := false
		for _, operation := range patch.Operations {
			pathMatches := false
			for _, path := range permission.Paths {
				g, err := glob.Compile(path)
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

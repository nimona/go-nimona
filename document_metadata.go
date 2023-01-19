package nimona

import (
	"github.com/gobwas/glob"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type (
	Metadata struct {
		Owner       string
		Permissions []Permissions
		Signature   Signature
		Timestamp   cbg.CborTime
	}
	Permissions struct {
		Paths      []string
		Operations PermissionsAllow
		Conditions PermissionsCondition
	}
	PermissionsAllow struct {
		// generic ops
		Read bool
		// json patch
		Add     bool
		Remove  bool
		Replace bool
		Move    bool
		Copy    bool
		Test    bool
	}
	PermissionsCondition struct {
		IsOwner bool
	}
)

// nolint: gocyclo
func CheckPermissions(metadata *Metadata, patch *StreamPatch) bool {
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

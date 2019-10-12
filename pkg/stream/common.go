package stream

import (
	json "encoding/json"

	crypto "nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

type (
	common struct {
		Context   string            `json:"@type:s,omitempty"`
		Stream    *object.Hash      `json:"stream:o,omitempty"`
		Parents   []*object.Hash    `json:"parents:ao,omitempty"`
		Policies  []*Policy         `json:"policies:ao,omitempty"`
		Signature *crypto.Signature `json:"@signature:o,omitempty"`
		Authors   []*Author         `json:"authors:ao"`
	}
)

func toCommon(o object.Object) *common {
	c := &common{}
	b, _ := json.Marshal(map[string]interface{}(o)) // nolint: errcheck
	json.Unmarshal(b, c)                            // nolint: errcheck
	return c
}

func Parents(o object.Object) []*object.Hash {
	return toCommon(o).Parents
}

func Stream(o object.Object) *object.Hash {
	return toCommon(o).Stream
}

func Policies(o object.Object) []*Policy {
	return toCommon(o).Policies
}

func Authors(o object.Object) []*Author {
	return toCommon(o).Authors
}

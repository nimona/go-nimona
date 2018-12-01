package crypto

import (
	"errors"
)

//go:generate go run nimona.io/go/cmd/objectify -schema /mandate -type Mandate -out mandate_generated.go

// Mandate to give authority to a aubject to perform certain actions on the
// authority's behalf
type Mandate struct {
	Authority   *Key       `json:"authority"`
	Subject     *Key       `json:"subject"`
	Description string     `json:"description"`
	Resources   []string   `json:"resources"`
	Actions     []string   `json:"actions"`
	Effect      string     `json:"effect"`
	Signature   *Signature `json:"@sig:O"`
}

// NewMandate returns a signed mandate given an authority key, a subject key,
// and a policy
func NewMandate(authority, subject *Key, description string, resources, actions []string, effect string) (*Mandate, error) {
	if authority == nil {
		return nil, errors.New("missing authority")
	}

	if subject == nil {
		return nil, errors.New("missing subject")
	}

	m := &Mandate{
		Authority:   authority,
		Subject:     subject,
		Description: description,
		Resources:   resources,
		Actions:     actions,
		Effect:      effect,
	}

	// o, err := encoding.NewObjectFromStruct(m)
	// if err != nil {
	// 	return nil, err
	// }

	// s, err := Sign(o, authority)
	// if err != nil {
	// 	return nil, err
	// }

	return m, nil
}

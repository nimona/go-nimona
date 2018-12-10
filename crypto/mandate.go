package crypto

import (
	"errors"
)

//go:generate go run nimona.io/go/cmd/objectify -schema /mandate -type Mandate -out mandate_generated.go

// Mandate to give authority to a aubject to perform certain actions on the
// authority's behalf
type Mandate struct {
	Signer      *Key       `json:"@signer"`
	Subject     *Key       `json:"subject"`
	Description string     `json:"description"`
	Resources   []string   `json:"resources"`
	Actions     []string   `json:"actions"`
	Effect      string     `json:"effect"`
	Signature   *Signature `json:"@signature"`
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
		Subject:     subject.GetPublicKey(),
		Description: description,
		Resources:   resources,
		Actions:     actions,
		Effect:      effect,
	}

	o := m.ToObject()
	if err := Sign(o, authority); err != nil {
		return nil, err
	}

	if err := m.FromObject(o); err != nil {
		return nil, err
	}

	return m, nil
}

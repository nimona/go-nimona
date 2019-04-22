package crypto

// import (
// 	"errors"
// )

// // Mandate to give signer to a aubject to perform certain actions on the
// // signer's behalf
// type Mandate struct {
// 	Signer      *PublicKey `json:"@signer"`
// 	Subject     *PublicKey `json:"subject"`
// 	Description string     `json:"description"`
// 	Resources   []string   `json:"resources"`
// 	Actions     []string   `json:"actions"`
// 	Effect      string     `json:"effect"`
// 	Signature   *Signature `json:"@signature"`
// }

// // NewMandate returns a signed mandate given a signer key, a subject key,
// // and a policy
// func NewMandate(signer *PrivateKey, subject *PublicKey, description string, resources, actions []string, effect string) (*Mandate, error) {
// 	if signer == nil {
// 		return nil, errors.New("missing signer")
// 	}

// 	if subject == nil {
// 		return nil, errors.New("missing subject")
// 	}

// 	m := &Mandate{
// 		Subject:     subject,
// 		Description: description,
// 		Resources:   resources,
// 		Actions:     actions,
// 		Effect:      effect,
// 	}

// 	o := m.ToObject()
// 	if err := Sign(o, signer); err != nil {
// 		return nil, err
// 	}

// 	if err := m.FromObject(o); err != nil {
// 		return nil, err
// 	}

// 	return m, nil
// }

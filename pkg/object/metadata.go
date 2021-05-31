package object

import "nimona.io/pkg/crypto"

// Metadata for object
type Metadata struct {
	Owner     crypto.PublicKey `nimona:"owner:s"`
	Datetime  string           `nimona:"datetime:s"`
	Parents   Parents          `nimona:"parents:m"`
	Policies  Policies         `nimona:"policies:am"`
	Stream    CID              `nimona:"stream:s"`
	Signature Signature        `nimona:"signature:m"`
}

func (m *Metadata) MarshalMap() (Map, error) {
	r := Map{}
	if !m.Owner.IsEmpty() {
		r["owner"] = String(m.Owner.String())
	}
	if len(m.Parents) > 0 {
		rv := Map{}
		for mk, mv := range m.Parents {
			rv[mk] = CIDArray(mv)
		}
		r["parents"] = rv
	}
	if len(m.Policies) > 0 {
		r["policies"] = m.Policies.Value()
	}
	if m.Stream != "" {
		r["stream"] = m.Stream
	}
	if m.Datetime != "" {
		r["datetime"] = String(m.Datetime)
	}
	if !m.Signature.IsEmpty() {
		v, err := m.Signature.MarshalMap()
		if err != nil {
			return nil, err
		}
		r["_signature"] = v
	}
	return r, nil
}

func (m *Metadata) UnmarshalMap(in Map) error {
	if t, ok := in["owner"]; ok {
		if s, ok := t.(String); ok {
			k := crypto.PublicKey{}
			err := k.UnmarshalString(string(s))
			if err != nil {
				return err
			}
			m.Owner = k
		}
	}
	if t, ok := in["datetime"]; ok {
		if s, ok := t.(String); ok {
			m.Datetime = string(s)
		}
	}
	if t, ok := in["parents"]; ok {
		if s, ok := t.(Map); ok {
			p := Parents{}
			for mk, mv := range s {
				ma, ok := mv.(CIDArray)
				if !ok {
					continue
				}
				p[mk] = ma
			}
			m.Parents = p
		}
	}
	if t, ok := in["policies"]; ok {
		if s, ok := t.(MapArray); ok {
			m.Policies = PoliciesFromValue(s)
		}
	}
	if t, ok := in["stream"]; ok {
		if s, ok := t.(CID); ok {
			m.Stream = s
		}
	}
	if t, ok := in["_signature"]; ok {
		if s, ok := t.(Map); ok {
			err := m.Signature.UnmarshalMap(s)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

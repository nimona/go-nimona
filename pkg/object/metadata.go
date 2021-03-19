package object

import "nimona.io/pkg/crypto"

// Metadata for object
type Metadata struct {
	Owner     *crypto.PublicKey
	Datetime  string
	Parents   Parents
	Policies  Policies
	Stream    CID
	Signature Signature
}

func (m Metadata) Map() Map {
	r := Map{}
	if m.Owner != nil {
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
		r["_signature"] = m.Signature.Map()
	}
	return r
}

func MetadataFromMap(s Map) Metadata {
	r := Metadata{}
	if t, ok := s["owner"]; ok {
		if s, ok := t.(String); ok {
			k := &crypto.PublicKey{}
			if err := k.UnmarshalString(string(s)); err == nil {
				r.Owner = k
			}
		}
	}
	if t, ok := s["datetime"]; ok {
		if s, ok := t.(String); ok {
			r.Datetime = string(s)
		}
	}
	if t, ok := s["parents"]; ok {
		if s, ok := t.(Map); ok {
			p := Parents{}
			for mk, mv := range s {
				ma, ok := mv.(CIDArray)
				if !ok {
					continue
				}
				p[mk] = ma
			}
			r.Parents = p
		}
	}
	if t, ok := s["policies"]; ok {
		if s, ok := t.(MapArray); ok {
			r.Policies = PoliciesFromValue(s)
		}
	}
	if t, ok := s["stream"]; ok {
		if s, ok := t.(CID); ok {
			r.Stream = s
		}
	}
	if t, ok := s["_signature"]; ok {
		if s, ok := t.(Map); ok {
			r.Signature = SignatureFromMap(s)
		}
	}

	return r
}

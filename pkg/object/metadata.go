package object

import "nimona.io/pkg/crypto"

// Metadata for object
type Metadata struct {
	Owner     crypto.PublicKey
	Datetime  string
	Parents   []CID
	Policy    Policy
	Stream    CID
	Signature Signature
}

func (m Metadata) Map() Map {
	r := Map{}
	if !m.Owner.IsEmpty() {
		r["owner"] = String(m.Owner)
	}
	if len(m.Parents) > 0 {
		rv := make(StringArray, len(m.Parents))
		for i, v := range m.Parents {
			rv[i] = String(v)
		}
		r["parents"] = rv
	}
	if len(m.Policy.Actions) > 0 {
		r["policy"] = m.Policy.Map()
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
			r.Owner = crypto.PublicKey(s)
		}
	}
	if t, ok := s["datetime"]; ok {
		if s, ok := t.(String); ok {
			r.Datetime = string(s)
		}
	}
	if t, ok := s["parents"]; ok {
		if s, ok := t.(StringArray); ok {
			hs := make([]CID, len(s))
			for i, h := range s {
				hs[i] = CID(h)
			}
			r.Parents = hs
		}
	}
	if t, ok := s["policy"]; ok {
		if s, ok := t.(Map); ok {
			r.Policy = PolicyFromMap(s)
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

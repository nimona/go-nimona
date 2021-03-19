package object

import (
	"nimona.io/pkg/crypto"
)

type (
	PolicyType       string
	PolicyAction     string
	PolicyEffect     string
	EvaluationResult string

	// Policy for object metadata
	Policy struct {
		Name      string
		Type      PolicyType
		Subjects  []crypto.PublicKey
		Resources []string
		Actions   []PolicyAction
		Effect    PolicyEffect
	}

	// Policies
	Policies []Policy

	// evaluation state
	evaluation struct {
		// target
		subject  crypto.PublicKey
		resource string
		action   PolicyAction
		// result
		explicitMatches int
		effect          EvaluationResult
	}
)

const (
	// Policy types
	SignaturePolicy PolicyType = "signature"

	// Policy actions
	ReadAction PolicyAction = "read"

	// Policy effects
	AllowEffect PolicyEffect = "allow"
	DenyEffect  PolicyEffect = "deny"

	// Policy Evaluation results
	Deny  EvaluationResult = "deny"
	Allow EvaluationResult = "allow"
)

func (ps Policies) Value() MapArray {
	a := make(MapArray, len(ps))
	for i, p := range ps {
		a[i] = p.Map()
	}
	return a
}

func (p Policy) Evaluate(
	subject crypto.PublicKey,
	resource string,
	action PolicyAction,
) EvaluationResult {
	e := &evaluation{
		subject:  subject,
		resource: resource,
		action:   action,
		effect:   Allow,
	}
	e.Process(p)
	return e.effect
}

func (ps Policies) Evaluate(
	subject crypto.PublicKey,
	resource string,
	action PolicyAction,
) EvaluationResult {
	e := &evaluation{
		subject:  subject,
		resource: resource,
		action:   action,
		effect:   Allow,
	}
	for _, p := range ps {
		e.Process(p)
	}
	return e.effect
}

func (e *evaluation) Process(
	p Policy,
) {
	explicitMatches := 0
	subjectMatches := len(p.Subjects) == 0
	if len(p.Subjects) > 0 {
		for _, s := range p.Subjects {
			if e.subject.Equals(s) {
				explicitMatches++
				subjectMatches = true
				break
			}
		}
	}

	resourceMatches := len(p.Resources) == 0
	if len(p.Resources) > 0 {
		for _, s := range p.Resources {
			if e.resource == s {
				explicitMatches++
				resourceMatches = true
				break
			}
		}
	}

	actionMatches := len(p.Actions) == 0
	if len(p.Actions) > 0 {
		for _, s := range p.Actions {
			if e.action == s {
				explicitMatches++
				actionMatches = true
				break
			}
		}
	}

	policyMatched := subjectMatches && resourceMatches && actionMatches
	if policyMatched {
		if explicitMatches >= e.explicitMatches {
			e.effect = EvaluationResult(p.Effect)
			e.explicitMatches = explicitMatches
		}
	}
}

func (p Policy) Map() Map {
	r := Map{}
	if p.Type != "" {
		r["type"] = String(p.Type)
	}
	if len(p.Subjects) > 0 {
		ss := make(StringArray, len(p.Subjects))
		for i, s := range p.Subjects {
			ss[i] = String(s.String())
		}
		r["subjects"] = ss
	}
	if len(p.Resources) > 0 {
		ss := make(StringArray, len(p.Resources))
		for i, s := range p.Resources {
			ss[i] = String(s)
		}
		r["resources"] = ss
	}
	if len(p.Actions) > 0 {
		ss := make(StringArray, len(p.Actions))
		for i, s := range p.Actions {
			ss[i] = String(s)
		}
		r["actions"] = ss
	}
	if p.Effect != "" {
		r["effect"] = String(p.Effect)
	}
	return r
}

func PoliciesFromValue(a MapArray) Policies {
	p := make(Policies, len(a))
	for i, m := range a {
		p[i] = PolicyFromMap(m)
	}
	return p
}

func PolicyFromMap(m Map) Policy {
	r := Policy{}
	if t, ok := m["type"]; ok {
		if s, ok := t.(String); ok {
			r.Type = PolicyType(s)
		}
	}
	if t, ok := m["subjects"]; ok {
		if s, ok := t.(StringArray); ok {
			p := make([]crypto.PublicKey, len(s))
			for i, v := range s {
				k := &crypto.PublicKey{}
				if err := k.UnmarshalString(string(v)); err == nil {
					p[i] = *k
				}
			}
			r.Subjects = p
		}
	}
	if t, ok := m["resources"]; ok {
		if s, ok := t.(StringArray); ok {
			r.Resources = FromStringArray(s)
		}
	}
	if t, ok := m["actions"]; ok {
		if s, ok := t.(StringArray); ok {
			p := make([]PolicyAction, len(s))
			for i, v := range s {
				p[i] = PolicyAction(v)
			}
			r.Actions = p
		}
	}
	if t, ok := m["effect"]; ok {
		if s, ok := t.(String); ok {
			r.Effect = PolicyEffect(s)
		}
	}
	return r
}

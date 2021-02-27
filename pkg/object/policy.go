package object

import "nimona.io/pkg/crypto"

type (
	PolicyType       string
	PolicyAction     string
	PolicyEffect     string
	EvaluationResult string

	// Policy for object metadata
	Policy struct {
		Type      PolicyType
		Subjects  []crypto.PublicKey
		Resources []string
		Actions   []PolicyAction
		Effect    PolicyEffect
	}

	// Policies
	Policies []Policy
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
	ExplicitDeny  EvaluationResult = "ExplicitDeny"
	ImplicitDeny  EvaluationResult = "ImplicitDeny"
	ExplicitAllow EvaluationResult = "ExplicitAllow"
)

func (ps Policies) Value() MapArray {
	a := make(MapArray, len(ps))
	for i, p := range ps {
		a[i] = p.Map()
	}
	return a
}

func (ps Policies) Evaluate(
	subject crypto.PublicKey,
	resource string,
	action PolicyAction,
) EvaluationResult {
	if len(ps) == 0 {
		return ExplicitAllow
	}
	allowed := false
	for _, p := range ps {
		r := p.Evaluate(
			subject,
			resource,
			action,
		)
		if r == ExplicitDeny {
			return ExplicitDeny
		}
		if r == ExplicitAllow {
			allowed = true
		}
	}
	if allowed {
		return ExplicitAllow
	}
	return ImplicitDeny
}

func (p Policy) Evaluate(
	subject crypto.PublicKey,
	resource string,
	action PolicyAction,
) EvaluationResult {
	subjectMatches := false
	resourceMatches := false
	actionMatches := false
	if len(p.Subjects) == 0 {
		subjectMatches = true
	} else {
		for _, s := range p.Subjects {
			if subject.Equals(s) {
				subjectMatches = true
				break
			}
		}
	}
	if len(p.Resources) == 0 {
		resourceMatches = true
	} else {
		for _, s := range p.Resources {
			if resource == s {
				resourceMatches = true
				break
			}
		}
	}
	if len(p.Actions) == 0 {
		actionMatches = true
	} else {
		for _, s := range p.Actions {
			if action == s {
				actionMatches = true
				break
			}
		}
	}
	if subjectMatches && resourceMatches && actionMatches {
		if p.Effect == AllowEffect {
			return ExplicitAllow
		}
		return ExplicitDeny
	}
	return ImplicitDeny
}

func (p Policy) Map() Map {
	r := Map{}
	if p.Type != "" {
		r["type"] = String(p.Type)
	}
	if len(p.Subjects) > 0 {
		ss := make(StringArray, len(p.Subjects))
		for i, s := range p.Subjects {
			ss[i] = String(s)
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
				p[i] = crypto.PublicKey(v)
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

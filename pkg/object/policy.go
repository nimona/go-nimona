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
		Name      string             `nimona:"name:s"`
		Type      PolicyType         `nimona:"type:s"`
		Subjects  []crypto.PublicKey `nimona:"subjects:as"`
		Resources []string           `nimona:"resources:as"`
		Actions   []PolicyAction     `nimona:"actions:as"`
		Effect    PolicyEffect       `nimona:"effect:s"`
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

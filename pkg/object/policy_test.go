package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
)

func TestPolicies_Evaluate(t *testing.T) {
	k0, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	k1, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	p0 := k0.PublicKey()
	p1 := k1.PublicKey()
	type args struct {
		subject  crypto.PublicKey
		resource string
		action   PolicyAction
	}
	tests := []struct {
		name     string
		policies Policies
		args     args
		want     EvaluationResult
	}{{
		name: "valid, all match",
		args: args{
			subject:  p0,
			action:   ReadAction,
			resource: "foo",
		},
		policies: Policies{{
			Type: SignaturePolicy,
			Subjects: []crypto.PublicKey{
				p0,
			},
			Actions: []PolicyAction{
				ReadAction,
			},
			Resources: []string{
				"foo",
				"bar",
			},
			Effect: AllowEffect,
		}},
		want: ExplicitAllow,
	}, {
		name: "valid, no resource",
		args: args{
			subject:  p0,
			action:   ReadAction,
			resource: "something",
		},
		policies: Policies{{
			Type: SignaturePolicy,
			Subjects: []crypto.PublicKey{
				p0,
			},
			Actions: []PolicyAction{
				ReadAction,
			},
			Effect: AllowEffect,
		}},
		want: ExplicitAllow,
	}, {
		name: "valid, no subject",
		args: args{
			subject:  p0,
			action:   ReadAction,
			resource: "something",
		},
		policies: Policies{{
			Type: SignaturePolicy,
			Actions: []PolicyAction{
				ReadAction,
			},
			Resources: []string{"something"},
			Effect:    AllowEffect,
		}},
		want: ExplicitAllow,
	}, {
		name: "valid, no action",
		args: args{
			subject:  p0,
			action:   ReadAction,
			resource: "something",
		},
		policies: Policies{{
			Type: SignaturePolicy,
			Subjects: []crypto.PublicKey{
				p0,
			},
			Resources: []string{"something"},
			Effect:    AllowEffect,
		}},
		want: ExplicitAllow,
	}, {
		name: "valid, no subject, no subject",
		args: args{
			subject:  p0,
			action:   ReadAction,
			resource: "something",
		},
		policies: Policies{{
			Type: SignaturePolicy,
			Actions: []PolicyAction{
				ReadAction,
			},
			Effect: AllowEffect,
		}},
		want: ExplicitAllow,
	}, {
		name: "valid, no subject, no subject, no action",
		args: args{
			subject:  p0,
			action:   ReadAction,
			resource: "something",
		},
		policies: Policies{{
			Type:   SignaturePolicy,
			Effect: AllowEffect,
		}},
		want: ExplicitAllow,
	}, {
		name: "invalid, no resource, subject doesn't match",
		args: args{
			subject: p1,
			action:  ReadAction,
		},
		policies: Policies{{
			Type: SignaturePolicy,
			Subjects: []crypto.PublicKey{
				p0,
			},
			Actions: []PolicyAction{
				ReadAction,
			},
			Effect: AllowEffect,
		}},
		want: ImplicitDeny,
	}, {
		name: "invalid, no resource, action doesn't match",
		args: args{
			subject: p0,
			action:  PolicyAction("foo"),
		},
		policies: Policies{{
			Type: SignaturePolicy,
			Subjects: []crypto.PublicKey{
				p0,
			},
			Actions: []PolicyAction{
				ReadAction,
			},
			Effect: AllowEffect,
		}},
		want: ImplicitDeny,
	}, {
		name: "invalid, resource doesn't match",
		args: args{
			subject:  p0,
			action:   ReadAction,
			resource: "foo",
		},
		policies: Policies{{
			Type: SignaturePolicy,
			Subjects: []crypto.PublicKey{
				p0,
			},
			Actions: []PolicyAction{
				ReadAction,
			},
			Resources: []string{
				"not-foo",
				"bar",
			},
			Effect: AllowEffect,
		}},
		want: ImplicitDeny,
	}, {
		name: "invalid, no resource, deny",
		args: args{
			subject: p0,
			action:  ReadAction,
		},
		policies: Policies{{
			Type: SignaturePolicy,
			Subjects: []crypto.PublicKey{
				p0,
			},
			Actions: []PolicyAction{
				ReadAction,
			},
			Effect: DenyEffect,
		}},
		want: ExplicitDeny,
	}, {
		name: "invalid, no resource, deny",
		args: args{
			subject: p0,
			action:  ReadAction,
		},
		policies: Policies{{
			Type: SignaturePolicy,
			Subjects: []crypto.PublicKey{
				p0,
			},
			Actions: []PolicyAction{
				ReadAction,
			},
			Effect: DenyEffect,
		}},
		want: ExplicitDeny,
	}, {
		name: "multiple, invalid, allow all, deny one",
		args: args{
			subject:  p0,
			action:   ReadAction,
			resource: "foo",
		},
		policies: Policies{{
			Type:   SignaturePolicy,
			Effect: AllowEffect,
		}, {
			Type:     SignaturePolicy,
			Subjects: []crypto.PublicKey{p0},
			Effect:   DenyEffect,
		}},
		want: ExplicitDeny,
	}, {
		name: "multiple, valid, allow all, deny one",
		args: args{
			subject:  p1,
			action:   ReadAction,
			resource: "foo",
		},
		policies: Policies{{
			Type:   SignaturePolicy,
			Effect: AllowEffect,
		}, {
			Type:     SignaturePolicy,
			Subjects: []crypto.PublicKey{p0},
			Effect:   DenyEffect,
		}},
		want: ExplicitAllow,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.policies.Evaluate(
				tt.args.subject,
				tt.args.resource,
				tt.args.action,
			)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPolicy_Map(t *testing.T) {
	k0, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	k1, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	p0 := k0.PublicKey()
	p1 := k1.PublicKey()
	p := Policy{
		Type:      SignaturePolicy,
		Subjects:  []crypto.PublicKey{p0, p1},
		Resources: []string{"foo", "bar"},
		Actions:   []PolicyAction{ReadAction, "foo", "bar"},
		Effect:    AllowEffect,
	}
	m := p.Map()
	g := PolicyFromMap(m)
	require.Equal(t, p, g)
}

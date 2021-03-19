package object

import (
	"fmt"
	"os"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
)

func TestPolicy_Evaluate_Table1(t *testing.T) {
	s0, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	s1, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	policies := Policies{{
		Name:      "#0 - allow s* r* a*",
		Effect:    AllowEffect,
		Subjects:  nil,
		Resources: nil,
		Actions:   nil,
	}, {
		Name:      "#1 - deny s* r* a1",
		Effect:    DenyEffect,
		Subjects:  nil,
		Resources: nil,
		Actions:   []PolicyAction{"a1"},
	}, {
		Name:      "#2 - deny s0 r* a*",
		Effect:    DenyEffect,
		Subjects:  []crypto.PublicKey{*s0.PublicKey()},
		Resources: nil,
		Actions:   nil,
	}, {
		Name:      "#3 - allow s0 r* a2",
		Effect:    AllowEffect,
		Subjects:  []crypto.PublicKey{*s0.PublicKey()},
		Resources: nil,
		Actions:   []PolicyAction{"a2"},
	}, {
		Name:      "#4 - allow s0 r0 a1",
		Effect:    AllowEffect,
		Subjects:  []crypto.PublicKey{*s0.PublicKey()},
		Resources: []string{"r0"},
		Actions:   []PolicyAction{"a1"},
	}, {
		Name:      "#5 - deny s* r0 a2",
		Effect:    DenyEffect,
		Subjects:  nil,
		Resources: []string{"r0"},
		Actions:   []PolicyAction{"a2"},
	}, {
		Name:      "#6 - deny s* r* a*",
		Effect:    AllowEffect,
		Subjects:  nil,
		Resources: nil,
		Actions:   nil,
	}}

	runTest := func(
		t *testing.T,
		targetName string,
		subject crypto.PublicKey,
		resource string,
		action PolicyAction,
		expectations []EvaluationResult,
	) {
		t.Helper()
		require.Equal(t, len(policies), len(expectations))
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"Name",
			"Subject",
			"Resource",
			"Action",
			targetName,
		})
		for i, p := range policies {
			ps := policies[:i+1]
			got := ps.Evaluate(
				subject,
				resource,
				action,
			)
			table.Append([]string{
				p.Name,
				fmt.Sprintf("%v", p.Subjects),
				fmt.Sprintf("%v", p.Resources),
				fmt.Sprintf("%v", p.Actions),
				fmt.Sprintf("%v", got),
			})
			assert.Equal(t, expectations[i], got)
		}
		table.Render()
	}

	// s0, r0, a1
	runTest(t, "s0, r0, a1", *s0.PublicKey(), "r0", "a1", []EvaluationResult{
		Allow,
		Deny,
		Deny,
		Deny,
		Allow,
		Allow,
		Allow,
	})

	// s0, r0, a2
	runTest(t, "s0, r0, a2", *s0.PublicKey(), "r0", "a2", []EvaluationResult{
		Allow,
		Allow,
		Deny,
		Allow,
		Allow,
		Deny,
		Deny,
	})

	// s1, r0, a1
	runTest(t, "s1, r0, a1", *s1.PublicKey(), "r0", "a1", []EvaluationResult{
		Allow,
		Deny,
		Deny,
		Deny,
		Deny,
		Deny,
		Deny,
	})

	// s1, r0, a2
	runTest(t, "s1, r0, a2", *s1.PublicKey(), "r0", "a2", []EvaluationResult{
		Allow,
		Allow,
		Allow,
		Allow,
		Allow,
		Deny,
		Deny,
	})
}

func TestPolicy_Evaluate_Table2(t *testing.T) {
	s0, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	s1, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	policies := Policies{{
		Name:      "#0 - allow s1 r0 a1",
		Effect:    AllowEffect,
		Subjects:  []crypto.PublicKey{*s1.PublicKey()},
		Resources: []string{"r0"},
		Actions:   []PolicyAction{"a1"},
	}, {
		Name:      "#1 - deny s* r* a*",
		Effect:    DenyEffect,
		Subjects:  nil,
		Resources: nil,
		Actions:   nil,
	}, {
		Name:      "#2 - allow s* r* a*",
		Effect:    AllowEffect,
		Subjects:  nil,
		Resources: nil,
		Actions:   nil,
	}, {
		Name:      "#3 - deny s0 r* a*",
		Effect:    DenyEffect,
		Subjects:  []crypto.PublicKey{*s0.PublicKey()},
		Resources: nil,
		Actions:   nil,
	}, {
		Name:      "#4 - deny s* r0 a*",
		Effect:    DenyEffect,
		Subjects:  nil,
		Resources: []string{"r0"},
		Actions:   nil,
	}, {
		Name:      "#5 - deny s* r0 a2",
		Effect:    DenyEffect,
		Subjects:  nil,
		Resources: []string{"r0"},
		Actions:   []PolicyAction{"a2"},
	}, {
		Name:      "#6 - allow s* r* a2",
		Effect:    AllowEffect,
		Subjects:  nil,
		Resources: nil,
		Actions:   []PolicyAction{"a2"},
	}, {
		Name:      "#7 - allow s0 r0 a*",
		Effect:    AllowEffect,
		Subjects:  []crypto.PublicKey{*s0.PublicKey()},
		Resources: []string{"r0"},
		Actions:   nil,
	}}

	runTest := func(
		t *testing.T,
		targetName string,
		subject crypto.PublicKey,
		resource string,
		action PolicyAction,
		expectations []EvaluationResult,
	) {
		t.Helper()
		require.Equal(t, len(policies), len(expectations))
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"Name",
			"Subject",
			"Resource",
			"Action",
			targetName,
		})
		for i, p := range policies {
			ps := policies[:i+1]
			got := ps.Evaluate(
				subject,
				resource,
				action,
			)
			table.Append([]string{
				p.Name,
				fmt.Sprintf("%v", p.Subjects),
				fmt.Sprintf("%v", p.Resources),
				fmt.Sprintf("%v", p.Actions),
				fmt.Sprintf("%v", got),
			})
			assert.Equal(t, expectations[i], got)
		}
		table.Render()
	}

	// s0, r0, a1
	runTest(t, "s0, r0, a1", *s0.PublicKey(), "r0", "a1", []EvaluationResult{
		Allow,
		Deny,
		Allow,
		Deny,
		Deny,
		Deny,
		Deny,
		Allow,
	})

	// s0, r0, a2
	runTest(t, "s0, r0, a2", *s0.PublicKey(), "r0", "a2", []EvaluationResult{
		Allow,
		Deny,
		Allow,
		Deny,
		Deny,
		Deny,
		Deny,
		Allow,
	})

	// s1, r0, a1
	runTest(t, "s1, r0, a1", *s1.PublicKey(), "r0", "a1", []EvaluationResult{
		Allow,
		Allow,
		Allow,
		Allow,
		Allow,
		Allow,
		Allow,
		Allow,
	})

	// s1, r0, a2
	runTest(t, "s1, r0, a2", *s1.PublicKey(), "r0", "a2", []EvaluationResult{
		Allow,
		Deny,
		Allow,
		Allow,
		Deny,
		Deny,
		Deny,
		Deny,
	})
}

// func TestPolicies_Evaluate(t *testing.T) {
// 	k0, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
// 	require.NoError(t, err)
// 	k1, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
// 	require.NoError(t, err)
// 	p0 := k0.PublicKey()
// 	p1 := k1.PublicKey()
// 	type args struct {
// 		subject  crypto.PublicKey
// 		resource string
// 		action   PolicyAction
// 	}
// 	tests := []struct {
// 		name     string
// 		policies Policies
// 		args     args
// 		want     EvaluationResult
// 	}{{
// 		name: "valid, no policies",
// 		args: args{
// 			subject:  p0,
// 			action:   ReadAction,
// 			resource: "foo",
// 		},
// 		policies: Policies{},
// 		want:     Allow,
// 	}, {
// 		name: "valid, all match",
// 		args: args{
// 			subject:  p0,
// 			action:   ReadAction,
// 			resource: "foo",
// 		},
// 		policies: Policies{{
// 			Type: SignaturePolicy,
// 			Subjects: []crypto.PublicKey{
// 				p0,
// 			},
// 			Actions: []PolicyAction{
// 				ReadAction,
// 			},
// 			Resources: []string{
// 				"foo",
// 				"bar",
// 			},
// 			Effect: AllowEffect,
// 		}},
// 		want: Allow,
// 	}, {
// 		name: "valid, no resource",
// 		args: args{
// 			subject:  p0,
// 			action:   ReadAction,
// 			resource: "something",
// 		},
// 		policies: Policies{{
// 			Type: SignaturePolicy,
// 			Subjects: []crypto.PublicKey{
// 				p0,
// 			},
// 			Actions: []PolicyAction{
// 				ReadAction,
// 			},
// 			Effect: AllowEffect,
// 		}},
// 		want: Allow,
// 	}, {
// 		name: "valid, no subject",
// 		args: args{
// 			subject:  p0,
// 			action:   ReadAction,
// 			resource: "something",
// 		},
// 		policies: Policies{{
// 			Type: SignaturePolicy,
// 			Actions: []PolicyAction{
// 				ReadAction,
// 			},
// 			Resources: []string{"something"},
// 			Effect:    AllowEffect,
// 		}},
// 		want: Allow,
// 	}, {
// 		name: "valid, no action",
// 		args: args{
// 			subject:  p0,
// 			action:   ReadAction,
// 			resource: "something",
// 		},
// 		policies: Policies{{
// 			Type: SignaturePolicy,
// 			Subjects: []crypto.PublicKey{
// 				p0,
// 			},
// 			Resources: []string{"something"},
// 			Effect:    AllowEffect,
// 		}},
// 		want: Allow,
// 	}, {
// 		name: "valid, no subject, no subject",
// 		args: args{
// 			subject:  p0,
// 			action:   ReadAction,
// 			resource: "something",
// 		},
// 		policies: Policies{{
// 			Type: SignaturePolicy,
// 			Actions: []PolicyAction{
// 				ReadAction,
// 			},
// 			Effect: AllowEffect,
// 		}},
// 		want: Allow,
// 	}, {
// 		name: "valid, no subject, no subject, no action",
// 		args: args{
// 			subject:  p0,
// 			action:   ReadAction,
// 			resource: "something",
// 		},
// 		policies: Policies{{
// 			Type:   SignaturePolicy,
// 			Effect: AllowEffect,
// 		}},
// 		want: Allow,
// 	}, {
// 		name: "invalid, no resource, subject doesn't match",
// 		args: args{
// 			subject: p1,
// 			action:  ReadAction,
// 		},
// 		policies: Policies{{
// 			Type: SignaturePolicy,
// 			Subjects: []crypto.PublicKey{
// 				p0,
// 			},
// 			Actions: []PolicyAction{
// 				ReadAction,
// 			},
// 			Effect: AllowEffect,
// 		}},
// 		want: Deny,
// 	}, {
// 		name: "invalid, no resource, action doesn't match",
// 		args: args{
// 			subject: p0,
// 			action:  PolicyAction("foo"),
// 		},
// 		policies: Policies{{
// 			Type: SignaturePolicy,
// 			Subjects: []crypto.PublicKey{
// 				p0,
// 			},
// 			Actions: []PolicyAction{
// 				ReadAction,
// 			},
// 			Effect: AllowEffect,
// 		}},
// 		want: Deny,
// 	}, {
// 		name: "invalid, resource doesn't match",
// 		args: args{
// 			subject:  p0,
// 			action:   ReadAction,
// 			resource: "foo",
// 		},
// 		policies: Policies{{
// 			Type: SignaturePolicy,
// 			Subjects: []crypto.PublicKey{
// 				p0,
// 			},
// 			Actions: []PolicyAction{
// 				ReadAction,
// 			},
// 			Resources: []string{
// 				"not-foo",
// 				"bar",
// 			},
// 			Effect: AllowEffect,
// 		}},
// 		want: Deny,
// 	}, {
// 		name: "invalid, no resource, deny",
// 		args: args{
// 			subject: p0,
// 			action:  ReadAction,
// 		},
// 		policies: Policies{{
// 			Type: SignaturePolicy,
// 			Subjects: []crypto.PublicKey{
// 				p0,
// 			},
// 			Actions: []PolicyAction{
// 				ReadAction,
// 			},
// 			Effect: DenyEffect,
// 		}},
// 		want: Deny,
// 	}, {
// 		name: "invalid, no resource, deny",
// 		args: args{
// 			subject: p0,
// 			action:  ReadAction,
// 		},
// 		policies: Policies{{
// 			Type: SignaturePolicy,
// 			Subjects: []crypto.PublicKey{
// 				p0,
// 			},
// 			Actions: []PolicyAction{
// 				ReadAction,
// 			},
// 			Effect: DenyEffect,
// 		}},
// 		want: Deny,
// 	}, {
// 		name: "multiple, invalid, allow all, deny one",
// 		args: args{
// 			subject:  p0,
// 			action:   ReadAction,
// 			resource: "foo",
// 		},
// 		policies: Policies{{
// 			Type:   SignaturePolicy,
// 			Effect: AllowEffect,
// 		}, {
// 			Type:     SignaturePolicy,
// 			Subjects: []crypto.PublicKey{p0},
// 			Effect:   DenyEffect,
// 		}},
// 		want: Deny,
// 	}, {
// 		name: "multiple, valid, allow all, deny one",
// 		args: args{
// 			subject:  p1,
// 			action:   ReadAction,
// 			resource: "foo",
// 		},
// 		policies: Policies{{
// 			Type:   SignaturePolicy,
// 			Effect: AllowEffect,
// 		}, {
// 			Type:     SignaturePolicy,
// 			Subjects: []crypto.PublicKey{p0},
// 			Effect:   DenyEffect,
// 		}},
// 		want: Allow,
// 	}}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got := tt.policies.Evaluate(
// 				tt.args.subject,
// 				tt.args.resource,
// 				tt.args.action,
// 			)
// 			assert.Equal(t, tt.want, got)
// 		})
// 	}
// }

// func TestPolicy_Map(t *testing.T) {
// 	k0, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
// 	require.NoError(t, err)
// 	k1, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
// 	require.NoError(t, err)
// 	p0 := k0.PublicKey()
// 	p1 := k1.PublicKey()
// 	p := Policy{
// 		Type:      SignaturePolicy,
// 		Subjects:  []crypto.PublicKey{p0, p1},
// 		Resources: []string{"foo", "bar"},
// 		Actions:   []PolicyAction{ReadAction, "foo", "bar"},
// 		Effect:    AllowEffect,
// 	}
// 	m := p.Map()
// 	g := PolicyFromMap(m)
// 	require.Equal(t, p, g)
// }

func TestPolicies_Map(t *testing.T) {
	k0, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	k1, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	p0 := k0.PublicKey()
	p1 := k1.PublicKey()
	a := Policies{{
		Type:      SignaturePolicy,
		Subjects:  []crypto.PublicKey{*p0, *p1},
		Resources: []string{"foo", "bar"},
		Actions:   []PolicyAction{ReadAction, "foo", "bar"},
		Effect:    AllowEffect,
	}, {
		Type:      SignaturePolicy,
		Subjects:  []crypto.PublicKey{*p0},
		Resources: []string{"foo"},
		Actions:   []PolicyAction{ReadAction},
		Effect:    DenyEffect,
	}}
	m := a.Value()
	g := PoliciesFromValue(m)
	require.Equal(t, a, g)
}

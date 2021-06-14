package object

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/chore"
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
		Subjects:  []crypto.PublicKey{s0.PublicKey()},
		Resources: nil,
		Actions:   nil,
	}, {
		Name:      "#3 - allow s0 r* a2",
		Effect:    AllowEffect,
		Subjects:  []crypto.PublicKey{s0.PublicKey()},
		Resources: nil,
		Actions:   []PolicyAction{"a2"},
	}, {
		Name:      "#4 - allow s0 r0 a1",
		Effect:    AllowEffect,
		Subjects:  []crypto.PublicKey{s0.PublicKey()},
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
	runTest(t, "s0, r0, a1", s0.PublicKey(), "r0", "a1", []EvaluationResult{
		Allow,
		Deny,
		Deny,
		Deny,
		Allow,
		Allow,
		Allow,
	})

	// s0, r0, a2
	runTest(t, "s0, r0, a2", s0.PublicKey(), "r0", "a2", []EvaluationResult{
		Allow,
		Allow,
		Deny,
		Allow,
		Allow,
		Deny,
		Deny,
	})

	// s1, r0, a1
	runTest(t, "s1, r0, a1", s1.PublicKey(), "r0", "a1", []EvaluationResult{
		Allow,
		Deny,
		Deny,
		Deny,
		Deny,
		Deny,
		Deny,
	})

	// s1, r0, a2
	runTest(t, "s1, r0, a2", s1.PublicKey(), "r0", "a2", []EvaluationResult{
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
		Subjects:  []crypto.PublicKey{s1.PublicKey()},
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
		Subjects:  []crypto.PublicKey{s0.PublicKey()},
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
		Subjects:  []crypto.PublicKey{s0.PublicKey()},
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
	runTest(t, "s0, r0, a1", s0.PublicKey(), "r0", "a1", []EvaluationResult{
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
	runTest(t, "s0, r0, a2", s0.PublicKey(), "r0", "a2", []EvaluationResult{
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
	runTest(t, "s1, r0, a1", s1.PublicKey(), "r0", "a1", []EvaluationResult{
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
	runTest(t, "s1, r0, a2", s1.PublicKey(), "r0", "a2", []EvaluationResult{
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

func TestPolicy_Marshal(t *testing.T) {
	k0, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	k1, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	p0 := k0.PublicKey()
	p1 := k1.PublicKey()
	p := &Policy{
		Type:      SignaturePolicy,
		Subjects:  []crypto.PublicKey{p0, p1},
		Resources: []string{"foo", "bar"},
		Actions:   []PolicyAction{ReadAction, "foo", "bar"},
		Effect:    AllowEffect,
	}

	m, err := marshalStruct(chore.MapHint, reflect.ValueOf(p))
	require.NoError(t, err)

	g := &Policy{}
	err = unmarshalMapToStruct(chore.MapHint, m, reflect.ValueOf(g))
	require.NoError(t, err)
	require.Equal(t, p, g)
}

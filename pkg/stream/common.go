package stream

import (
	json "encoding/json"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/object"
)

var (
	typeStreamSubscribed   = (&Subscribed{}).GetType()
	typeStreamUnsubscribed = (&Unsubscribed{}).GetType()
)

type (
	common struct {
		Type      string            `json:"@ctx:s,omitempty"`
		Context   string            `json:"@type:s,omitempty"`
		Stream    *object.Hash      `json:"stream:o,omitempty"`
		Parents   []*object.Hash    `json:"parents:ao,omitempty"`
		Policies  []*Policy         `json:"policies:ao,omitempty"`
		Signature *crypto.Signature `json:"@signature:o,omitempty"`
		Identity  *crypto.PublicKey `json:"@identity:o"`
	}
)

func toCommon(o object.Object) *common {
	c := &common{}
	b, _ := json.Marshal(map[string]interface{}(o)) // nolint: errcheck
	json.Unmarshal(b, c)                            // nolint: errcheck
	return c
}

func Parents(o object.Object) []*object.Hash {
	return toCommon(o).Parents
}

func Stream(o object.Object) *object.Hash {
	return toCommon(o).Stream
}

func Policies(o object.Object) []*Policy {
	return toCommon(o).Policies
}

func Identity(o object.Object) *crypto.PublicKey {
	return toCommon(o).Identity
}

func GetStreamSubscribers(os []object.Object) []*crypto.PublicKey {
	subs := map[*crypto.PublicKey]bool{}
	for _, o := range os {
		switch o.GetType() {
		case typeStreamSubscribed:
			e := &Subscribed{}
			e.FromObject(o) // nolint: errcheck
			subs[e.Identity] = true
		case typeStreamUnsubscribed:
			e := &Unsubscribed{}
			e.FromObject(o) // nolint: errcheck
			subs[e.Identity] = false
		}
	}
	cleanSubs := []*crypto.PublicKey{}
	for k, ok := range subs {
		if !ok {
			continue
		}
		cleanSubs = append(cleanSubs, k)
	}
	return cleanSubs
}

func GetStreamTails(os []object.Object) []object.Object {
	hm := map[string]bool{} // map[hash]isParent
	om := map[string]object.Object{}
	for _, o := range os {
		h := hash.New(o).String()
		if _, ok := hm[h]; !ok {
			hm[h] = false
		}
		for _, p := range Parents(o) {
			hm[p.Compact()] = true
		}
		om[h] = o
	}

	os = []object.Object{}
	for h, isParent := range hm {
		if isParent == false {
			os = append(os, om[h])
		}
	}

	return os
}

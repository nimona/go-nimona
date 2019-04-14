// Code generated by nimona.io/tools/objectify. DO NOT EDIT.

// +build !generate

package dht

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

const (
	ProviderType = "nimona.io/dht/provider"
)

// ToObject returns a f12n object
func (s Provider) ToObject() *object.Object {
	o := object.New()
	o.SetType(ProviderType)
	if len(s.ObjectIDs) > 0 {
		o.SetRaw("objectIDs", s.ObjectIDs)
	}
	if s.Signer != nil {
		o.SetRaw("@signer", s.Signer)
	}
	if s.Authority != nil {
		o.SetRaw("@authority", s.Authority)
	}
	if s.Signature != nil {
		o.SetRaw("@signature", s.Signature)
	}
	return o
}

func anythingToAnythingForProvider(
	from interface{},
	to interface{},
) error {
	config := &mapstructure.DecoderConfig{
		Result:  to,
		TagName: "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	if err := decoder.Decode(from); err != nil {
		return err
	}

	return nil
}

// FromObject populates the struct from a f12n object
func (s *Provider) FromObject(o *object.Object) error {
	atoa := anythingToAnythingForProvider
	if err := atoa(o.GetRaw("objectIDs"), &s.ObjectIDs); err != nil {
		return err
	}
	s.RawObject = o
	if v, ok := o.GetRaw("@signer").(*crypto.Key); ok {
		s.Signer = v
	} else if v, ok := o.GetRaw("@signer").(map[string]interface{}); ok {
		s.Signer = &crypto.Key{}
		o := &object.Object{}
		if err := o.FromMap(v); err != nil {
			return err
		}
		s.Signer.FromObject(o)
	}
	if v, ok := o.GetRaw("@authority").(*crypto.Key); ok {
		s.Authority = v
	} else if v, ok := o.GetRaw("@authority").(map[string]interface{}); ok {
		s.Authority = &crypto.Key{}
		o := &object.Object{}
		if err := o.FromMap(v); err != nil {
			return err
		}
		s.Authority.FromObject(o)
	}
	if v, ok := o.GetRaw("@signature").(*crypto.Signature); ok {
		s.Signature = v
	} else if v, ok := o.GetRaw("@signature").(map[string]interface{}); ok {
		s.Signature = &crypto.Signature{}
		o := &object.Object{}
		if err := o.FromMap(v); err != nil {
			return err
		}
		s.Signature.FromObject(o)
	}

	return nil
}

// GetType returns the object's type
func (s Provider) GetType() string {
	return ProviderType
}

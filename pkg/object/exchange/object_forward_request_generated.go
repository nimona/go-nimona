// Code generated by nimona.io/tools/objectify. DO NOT EDIT.

// +build !generate

package exchange

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

const (
	ObjectForwardRequestType = "/object-forward-request"
)

// ToObject returns a f12n object
func (s ObjectForwardRequest) ToObject() *object.Object {
	o := object.New()
	o.SetType(ObjectForwardRequestType)
	if s.Recipient != "" {
		o.SetRaw("recipient", s.Recipient)
	}
	if s.FwObject != nil {
		o.SetRaw("fwObject", s.FwObject)
	}
	if s.Signature != nil {
		o.SetRaw("@signature", s.Signature)
	}
	return o
}

func anythingToAnythingForObjectForwardRequest(
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
func (s *ObjectForwardRequest) FromObject(o *object.Object) error {
	atoa := anythingToAnythingForObjectForwardRequest
	if err := atoa(o.GetRaw("recipient"), &s.Recipient); err != nil {
		return err
	}
	if v, ok := o.GetRaw("fwObject").(map[string]interface{}); ok {
		s.FwObject = &object.Object{}
		s.FwObject.FromMap(v)
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

	if ao, ok := interface{}(s).(interface{ afterFromObject() }); ok {
		ao.afterFromObject()
	}

	return nil
}

// GetType returns the object's type
func (s ObjectForwardRequest) GetType() string {
	return ObjectForwardRequestType
}

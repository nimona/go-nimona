// Code generated by nimona.io/tools/objectify. DO NOT EDIT.

// +build !generate

package exchange

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

const (
	ObjectResponseType = "/object-response"
)

// ToObject returns a f12n object
func (s ObjectResponse) ToObject() *object.Object {
	o := object.New()
	o.SetType(ObjectResponseType)
	if s.RequestID != "" {
		o.SetRaw("requestID", s.RequestID)
	}
	if s.RequestedObject != nil {
		o.SetRaw("requestedObject", s.RequestedObject)
	}
	if s.Sender != nil {
		o.SetRaw("sender", s.Sender)
	}
	if s.Signature != nil {
		o.SetRaw("@signature", s.Signature)
	}
	return o
}

// FromObject populates the struct from a f12n object
func (s *ObjectResponse) FromObject(o *object.Object) error {
	if v, ok := o.GetRaw("requestID").(string); ok {
		s.RequestID = v
	}
	if v, ok := o.GetRaw("requestedObject").(*object.Object); ok {
		s.RequestedObject = v
	}
	if v, ok := o.GetRaw("sender").(*crypto.Key); ok {
		s.Sender = v
	} else if v, ok := o.GetRaw("sender").(*object.Object); ok {
		s.Sender = &crypto.Key{}
		s.Sender.FromObject(v)
	}
	if v, ok := o.GetRaw("@signature").(*crypto.Signature); ok {
		s.Signature = v
	} else if v, ok := o.GetRaw("@signature").(*object.Object); ok {
		s.Signature = &crypto.Signature{}
		s.Signature.FromObject(v)
	}
	return nil
}

// GetType returns the object's type
func (s ObjectResponse) GetType() string {
	return ObjectResponseType
}

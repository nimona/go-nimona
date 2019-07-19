// Code generated by nimona.io/tools/objectify. DO NOT EDIT.

// +build !generate

package handshake

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

const (
	SynAckType = "/handshake.syn-ack"
)

// ToObject returns a f12n object
func (s SynAck) ToObject() *object.Object {
	o := object.New()
	o.SetType(SynAckType)
	if s.Nonce != "" {
		o.SetRaw("nonce", s.Nonce)
	}
	if s.Peer != nil {
		o.SetRaw("peer", s.Peer)
	}
	if s.Signature != nil {
		o.SetRaw("@signature", s.Signature)
	}
	return o
}

func anythingToAnythingForSynAck(
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
func (s *SynAck) FromObject(o *object.Object) error {
	atoa := anythingToAnythingForSynAck
	if err := atoa(o.GetRaw("nonce"), &s.Nonce); err != nil {
		return err
	}
	if v, ok := o.GetRaw("peer").(*peer.Peer); ok {
		s.Peer = v
	} else if v, ok := o.GetRaw("peer").(map[string]interface{}); ok {
		s.Peer = &peer.Peer{}
		o := &object.Object{}
		if err := o.FromMap(v); err != nil {
			return err
		}
		s.Peer.FromObject(o)
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
func (s SynAck) GetType() string {
	return SynAckType
}

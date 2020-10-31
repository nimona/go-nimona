// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package peer

import (
	object "nimona.io/pkg/object"
)

type (
	Peer struct {
		Metadata     object.Metadata       `nimona:"metadata:m,omitempty"`
		Version      int64                 `nimona:"version:i,omitempty"`
		Addresses    []string              `nimona:"addresses:as,omitempty"`
		Bloom        []int64               `nimona:"bloom:ai,omitempty"`
		ContentTypes []string              `nimona:"contentTypes:as,omitempty"`
		Certificates []*object.Certificate `nimona:"certificates:ao,omitempty"`
		Relays       []*Peer               `nimona:"relays:ao,omitempty"`
	}
	LookupRequest struct {
		Metadata object.Metadata `nimona:"metadata:m,omitempty"`
		Nonce    string          `nimona:"nonce:s,omitempty"`
		Bloom    []int64         `nimona:"bloom:ai,omitempty"`
	}
	LookupResponse struct {
		Metadata object.Metadata `nimona:"metadata:m,omitempty"`
		Nonce    string          `nimona:"nonce:s,omitempty"`
		Bloom    []int64         `nimona:"bloom:ai,omitempty"`
		Peers    []*Peer         `nimona:"peers:ao,omitempty"`
	}
)

func (e *Peer) Type() string {
	return "nimona.io/peer.Peer"
}

func (e Peer) ToObject() *object.Object {
	o, err := object.Encode(&e)
	if err != nil {
		panic(err)
	}
	return o
}

func (e *Peer) FromObject(o *object.Object) error {
	return object.Decode(o, e)
}

func (e *LookupRequest) Type() string {
	return "nimona.io/LookupRequest"
}

func (e LookupRequest) ToObject() *object.Object {
	o, err := object.Encode(&e)
	if err != nil {
		panic(err)
	}
	return o
}

func (e *LookupRequest) FromObject(o *object.Object) error {
	return object.Decode(o, e)
}

func (e *LookupResponse) Type() string {
	return "nimona.io/LookupResponse"
}

func (e LookupResponse) ToObject() *object.Object {
	o, err := object.Encode(&e)
	if err != nil {
		panic(err)
	}
	return o
}

func (e *LookupResponse) FromObject(o *object.Object) error {
	return object.Decode(o, e)
}

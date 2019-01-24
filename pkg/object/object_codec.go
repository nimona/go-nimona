package object

import (
	"github.com/ugorji/go/codec"
)

// CodecDecodeSelf helper for cbor/json unmarshaling
func (o *Object) CodecDecodeSelf(dec *codec.Decoder) {
	m := map[string]interface{}{}
	dec.MustDecode(&m)
	o.FromMap(m)
}

// CodecEncodeSelf helper for cbor/json marshaling
func (o *Object) CodecEncodeSelf(enc *codec.Encoder) {
	m := o.ToMap()
	enc.MustEncode(m)
}

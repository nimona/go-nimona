package peers // import "nimona.io/go/peers"

import (
	"github.com/mitchellh/mapstructure"
	ucodec "github.com/ugorji/go/codec"

	"nimona.io/go/primitives"
)

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	Addresses []string `mapstructure:"addresses"`
	Signature *primitives.Signature
}

func (pi *PeerInfo) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/peer.info",
		Payload: map[string]interface{}{
			"addresses": pi.Addresses,
		},
		Signature: pi.Signature,
	}
}

func (pi *PeerInfo) FromBlock(block *primitives.Block) {
	if err := mapstructure.Decode(block.Payload, pi); err != nil {
		panic(err)
	}
	pi.Signature = block.Signature
}

// CodecDecodeSelf helper for cbor unmarshaling
func (pi *PeerInfo) CodecDecodeSelf(dec *ucodec.Decoder) {
	b := &primitives.Block{}
	dec.MustDecode(b)
	pi.FromBlock(b)
}

// CodecEncodeSelf helper for cbor marshaling
func (pi *PeerInfo) CodecEncodeSelf(enc *ucodec.Encoder) {
	b := pi.Block()
	enc.MustEncode(b)
}

func (pi *PeerInfo) Thumbprint() string {
	return pi.Signature.Key.Thumbprint()
}

func (pi *PeerInfo) Address() string {
	return "peer:" + pi.Signature.Key.Thumbprint()
}

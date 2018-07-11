package wire

// import (
// 	"crypto/rand"

// 	"go.mozilla.org/cose"

// 	"github.com/nimona/go-nimona/peer"
// 	"github.com/ugorji/go/codec"
// )

// // Message for exchanging data via the wire
// type Message struct {
// 	Version int
// 	Headers struct {
// 		ContentType string
// 		Recipients  []string
// 	}
// 	Payload   []byte
// 	Signature []byte
// }

// func fromSignMessage(sm *cose.SignMessage) (*Message, error) {
// 	message := &Message{}
// 	message.Headers.ContentType = sm.Headers.Protected["content-type"].(string)
// 	message.Headers.Recipients = sm.Headers.Protected["recipients"].([]string)
// 	message.Payload = sm.Payload
// 	message.Signature = sm.Signatures[0].SignatureBytes
// 	return message, nil
// }

// func SignerFromPeer(secretPeerInfo *peer.SecretPeerInfo) (*cose.Signer, error) {
// 	// TODO get correct alg
// 	signer, err := cose.NewSignerFromKey(cose.PS256, secretPeerInfo.GetSecretKey())
// 	if err != nil {
// 		return nil, err
// 	}

// 	return signer, nil
// }

// func Sign(message *Message, signerPeerInfo *peer.SecretPeerInfo) ([]byte, error) {
// 	signMessage, err := toSignMessage(message)
// 	if err != nil {
// 		return nil, err
// 	}

// 	signer, err := SignerFromPeer(signerPeerInfo)
// 	if err != nil {
// 		return nil, err
// 	}

// 	signature := cose.NewSignature()
// 	signature.Headers.Protected["alg"] = "PS256"
// 	signMessage.AddSignature(signature)

// 	signMessage.Sign(rand.Reader, nil, []cose.Signer{*signer})
// 	if err != nil {
// 		return nil, err
// 	}

// 	messageBytes, err := cose.Marshal(signMessage)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return messageBytes, nil
// }

// func toSignMessage(message *Message) (*cose.SignMessage, error) {
// 	msg := cose.NewSignMessage()
// 	msg.Payload = message.Payload
// 	msg.Headers.Protected["content-type"] = message.Headers.ContentType
// 	msg.Headers.Protected["recipients"] = message.Headers.Recipients
// 	return &msg, nil
// }

// // func (m *Message) Headers.ContentType string {
// // 	return m.Headers.ContentType
// // }

// // DecodePayload decodes the message's payload according to the coded,
// // and stores the result in the value pointed to by r.
// func (h *Message) DecodePayload(r interface{}) error {
// 	dec := codec.NewDecoderBytes(h.Payload, &codec.CborHandle{})
// 	return dec.Decode(r)
// }

// // EncodePayload encodes the given value using the message's codec, and stores
// // the result in the message's payload.
// func (h *Message) EncodePayload(r interface{}) error {
// 	payloadBytes := []byte{}
// 	enc := codec.NewEncoderBytes(&payloadBytes, &codec.CborHandle{})
// 	if err := enc.Encode(r); err != nil {
// 		return err
// 	}
// 	h.Payload = payloadBytes
// 	return nil
// }

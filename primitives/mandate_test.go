package primitives

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/go/base58"
	"nimona.io/go/codec"
)

func TestMandateBlock(t *testing.T) {
	ep := &Mandate{
		Subject: &Key{
			Algorithm: "key-alg",
			KeyID:     "subject-kid",
		},
		Policy: Policy{
			Description: "description",
			Subjects: []string{
				"subject1",
				"subject2",
			},
			Actions: []string{
				"action1",
				"action2",
			},
			Effect: "effect",
		},
		Signature: &Signature{
			Key: &Key{
				Algorithm: "key-alg",
				KeyID:     "authority-kid",
			},
			Alg:       "sig-alg",
			Signature: []byte("sig-bytes"),
		},
	}

	b := ep.Block()
	bs, _ := Marshal(b)

	b2, err := Unmarshal(bs)
	assert.NoError(t, err)

	p := &Mandate{}
	p.FromBlock(b2)

	assert.Equal(t, ep, p)
}

func TestMandateSelfEncode(t *testing.T) {
	eb := &Mandate{
		Subject: &Key{
			Algorithm: "key-alg",
			KeyID:     "subject-kid",
		},
		Policy: Policy{
			Description: "description",
			Subjects: []string{
				"subject1",
				"subject2",
			},
			Actions: []string{
				"action1",
				"action2",
			},
			Effect: "effect",
		},
		Signature: &Signature{
			Key: &Key{
				Algorithm: "key-alg",
				KeyID:     "authority-kid",
			},
			Alg:       "sig-alg",
			Signature: []byte("sig-bytes"),
		},
	}

	bs, err := codec.Marshal(eb)
	assert.NoError(t, err)

	assert.Equal(t, "jJtMiQjq9BmYw1cMBFWBctXuC8U94hbX7ohnhPxHMnNGnC9SN94jT4QTh"+
		"hN1NY6TygW4XY8SiJ28KQkTww9fwfeZfMVNbXxPNkyRCuCS5vLPDSoVRu96maZx4bUkz7"+
		"cUEY8sBAHXecJZVzKVU2gFiok4G2YwZH1RcZwg7CtV99Tc9w76m2vmAbK9VxJmLvSnot9"+
		"RaFyNTsd2Gh6BhWi455Z9vPFxrnUqLMg653M6AsMiuTKkWYiHftDMWbufSSahiDvZSLxL"+
		"rpwVnnXYvR2sPZ2DvCLkRMV64C4nNRfyqDdNxXtxd9mELstzHcEMv29dnXQZLs9VRXaaL"+
		"e6twKFELx1yfiZRe49Nd9io6ahMGT4XtjYMJ2RQqcs3ju11uF2MUJ9zRGbNf1iFKeJ2th"+
		"ePqsexroFFneKtyLJysf6Au8CcBB2LkbQ1Mh9hxrJZKU4MiAuJ", base58.Encode(bs))

	b := &Mandate{}
	err = codec.Unmarshal(bs, b)
	assert.NoError(t, err)

	assert.Equal(t, eb, b)
}

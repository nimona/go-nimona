package hyperspace

import (
	"nimona.io/pkg/crypto"
)

//go:generate $GOBIN/objectify -schema nimona.io/discovery/hyperspace/bloom -type ContentHashesBloom -in bloom.go -out bloom_generated.go

type (
	// ContentHashesBloom -
	ContentHashesBloom struct {
		BloomFilter []int             `json:"bloomFilter:ai"`
		Signature   *crypto.Signature `json:"@signature:o"`
	}
)

func (ch *ContentHashesBloom) Bloom() []int {
	return []int(ch.BloomFilter)
}

package hyperspace

//go:generate $GOBIN/objectify -schema nimona.io/discovery/hyperspace/bloom.request -type ContentHashesBloomRequest -in bloom_request.go -out bloom_request_generated.go

// ContentHashesBloomRequest -
type ContentHashesBloomRequest struct {
	BloomFilter []int `json:"bloomFilter:ai"`
}

func (ch *ContentHashesBloomRequest) Bloom() []int {
	return []int(ch.BloomFilter)
}

package hyperspace

func (ch *ContentProviderUpdated) Bloom() []int64 {
	return []int64(ch.BloomFilter)
}

func (ch *ContentProviderRequested) Bloom() []int64 {
	return []int64(ch.BloomFilter)
}

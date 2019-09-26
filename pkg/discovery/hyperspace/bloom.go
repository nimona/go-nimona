package hyperspace

func (ch *Announced) Bloom() []int64 {
	return []int64(ch.AvailableContentBloom)
}

func (ch *Request) Bloom() []int64 {
	return []int64(ch.QueryContentBloom)
}

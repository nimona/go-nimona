package nimona

type DocumentMapper interface {
	DocumentMap() *DocumentMap
	FromDocumentMap(*DocumentMap) error
}

type MessageWrapper struct {
	Type string `nimona:"$type"`
}
